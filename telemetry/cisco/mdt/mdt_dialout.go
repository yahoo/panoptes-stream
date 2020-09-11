//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package mdt

import (
	"bytes"
	"context"
	"errors"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	dialout "github.com/cisco-ie/nx-telemetry-proto/mdt_dialout"
	mdt "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

// Dialout represents MDT dial-out.
type Dialout struct {
	ctx        context.Context
	cfg        config.Config
	dataChan   chan []byte
	outChan    telemetry.ExtDSChan
	logger     *zap.Logger
	metrics    map[string]status.Metrics
	pathOutput map[string]string

	sync.RWMutex
}

// NewDialout returns a new instance of MDT dial-out.
func NewDialout(ctx context.Context, cfg config.Config, outChan telemetry.ExtDSChan) *Dialout {
	m := &Dialout{
		ctx:        ctx,
		cfg:        cfg,
		outChan:    outChan,
		logger:     cfg.Logger(),
		dataChan:   make(chan []byte, 1000),
		pathOutput: make(map[string]string),
	}

	for _, sensor := range cfg.Sensors() {
		m.pathOutput[sensor.Subscription] = sensor.Output
	}

	return m
}

// Start creates workers and starts gRPC server.
func (m *Dialout) Start() error {
	var grpcSrvOpts []grpc.ServerOption

	tlsConf := m.cfg.Global().Dialout.TLSConfig
	conf := m.cfg.Global().Dialout.Services["cisco.mdt"]

	if conf.Addr == "" {
		return errors.New("address is empty")
	}

	if conf.Workers < 1 {
		conf.Workers = 2
	}

	for i := 0; i < conf.Workers; i++ {
		go m.worker()
	}

	ln, err := net.Listen("tcp", conf.Addr)
	if err != nil {
		return err
	}

	if tlsConf.Enabled {
		tlsConfig, err := secret.GetTLSServerConfig(&tlsConf)
		if err != nil {
			return err
		}

		creds := grpc.Creds(credentials.NewTLS(tlsConfig))
		grpcSrvOpts = append(grpcSrvOpts, creds)
	}

	srv := grpc.NewServer(grpcSrvOpts...)
	dialout.RegisterGRPCMdtDialoutServer(srv, m)
	go srv.Serve(ln)

	m.logger.Info("cisco.mdt.dialout", zap.String("address", conf.Addr), zap.Bool("tls", m.cfg.Global().Dialout.TLSConfig.Enabled))

	return nil
}

// Update updates path to output once the configuration changed.
func (m *Dialout) Update() {
	m.Lock()
	defer m.Unlock()

	m.pathOutput = make(map[string]string)
	for _, sensor := range m.cfg.Sensors() {
		m.pathOutput[sensor.Subscription] = sensor.Output
	}
}

// MdtDialout gets stream metrics and fan-out to workers.
func (m *Dialout) MdtDialout(stream dialout.GRPCMdtDialout_MdtDialoutServer) error {
	var buf *bytes.Buffer

	p, ok := peer.FromContext(stream.Context())

	if !ok {
		m.logger.Warn("cisco.mdt.dialout", zap.String("event", "connect"), zap.String("host", "peer address is unavailable"))
	} else {
		m.logger.Info("cisco.mtd.dialout", zap.String("event", "connect"), zap.String("peer", p.Addr.String()))
	}

	for {
		dialoutArgs, err := stream.Recv()
		if err != nil {
			return err
		}

		if dialoutArgs.TotalSize == 0 {
			m.dataChan <- dialoutArgs.Data
			continue
		}

		buf.Write(dialoutArgs.Data)
		if int32(buf.Len()) >= dialoutArgs.TotalSize {
			m.dataChan <- dialoutArgs.Data
			buf.Reset()
		}
	}
}

func (m *Dialout) worker() {
	var buf = new(bytes.Buffer)
	for {
		select {
		case d, ok := <-m.dataChan:
			if !ok {
				return
			}

			if err := m.datastore(buf, d); err != nil {
				m.logger.Error("cisco.mdt.dialout", zap.Error(err))
			}

		case <-m.ctx.Done():
			return
		}
	}
}

func (m *Dialout) datastore(buf *bytes.Buffer, data []byte) error {
	tm := &mdt.Telemetry{}
	err := proto.Unmarshal(data, tm)
	if err != nil {
		return err
	}

	m.handler(buf, tm)

	return nil
}

func (m *Dialout) handler(buf *bytes.Buffer, tm *mdt.Telemetry) {
	var (
		prefix, output string
		timestamp      uint64
		err            error
	)

	for _, gpbkv := range tm.DataGpbkv {
		output, err = m.getOutput(tm.GetSubscriptionIdStr())
		if err != nil {
			m.logger.Error("cisco.mdt.dialout", zap.Error(err))
		}

		timestamp = getTimestamp(gpbkv.Timestamp, tm.MsgTimestamp)

		labels := map[string]string{
			"subscriptionId": tm.GetSubscriptionIdStr(),
			"nodeId":         tm.GetNodeIdStr(),
			"path":           tm.GetEncodingPath(),
		}

		prefix = tm.GetEncodingPath()

		var key, content *mdt.TelemetryField
		for _, field := range gpbkv.Fields {
			if field.Name == "keys" {
				key = field
			} else if field.Name == "content" {
				content = field
			}
		}

		if key == nil || content == nil {
			m.logger.Warn("cisco.mdt.dialout", zap.String("msg", "invalid data"))
			continue
		}

		for _, subFiled := range key.Fields {
			getKeyLabels(labels, subFiled)
		}

		kv := make(map[string]interface{})
		for _, subFiled := range content.Fields {
			getKey(buf, kv, subFiled)
		}

		for key, value := range kv {
			dataStore := telemetry.DataStore{
				"prefix":    prefix,
				"labels":    labels,
				"timestamp": timestamp,
				"system_id": tm.GetNodeIdStr(),
				"key":       key,
				"value":     value,
			}

			select {
			case m.outChan <- telemetry.ExtDataStore{
				DS:     dataStore,
				Output: output,
			}:
			default:
				m.logger.Warn("cisco.mdt.dialout", zap.String("error", "dataset drop"))
			}
		}

		buf.Reset()
	}
}

func (m *Dialout) getOutput(sub string) (string, error) {
	if m.cfg.Global().Dialout.DefaultOutput != "" {
		return m.cfg.Global().Dialout.DefaultOutput, nil
	}

	m.RLock()
	defer m.RUnlock()
	if output, ok := m.pathOutput[sub]; ok {
		return output, nil
	}

	return "", errors.New("output not found")
}
