//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package mdt

import (
	"bytes"
	"context"
	"net"
	"os"
	"strconv"

	mdt "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	mdtGRPC "git.vzbuilders.com/marshadrad/panoptes/telemetry/cisco/proto"
)

const (
	gpb = iota + 2
	gpbkv
	json
)

var mdtVersion = "0.0.1"

// MDT represents Model-Driven Telemetry.
type MDT struct {
	conn          *grpc.ClientConn
	subscriptions []string

	dataChan chan []byte
	outChan  telemetry.ExtDSChan
	logger   *zap.Logger

	metrics    map[string]status.Metrics
	systemID   string
	pathOutput map[string]string
}

// New returns new instance of NMI.
func New(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	var metrics = make(map[string]status.Metrics)

	metrics["gRPCDataTotal"] = status.NewCounter("cisco_mdt_grpc_data_total", "")
	metrics["dropsTotal"] = status.NewCounter("cisco_mdt_drops_total", "")
	metrics["errorsTotal"] = status.NewCounter("cisco_mdt_errors_total", "")
	metrics["processNSecond"] = status.NewGauge("cisco_mdt_process_nanosecond", "")

	status.Register(status.Labels{"host": conn.Target()}, metrics)

	m := &MDT{
		conn:       conn,
		outChan:    outChan,
		logger:     logger,
		dataChan:   make(chan []byte, 1000),
		pathOutput: make(map[string]string),
	}

	for _, sensor := range sensors {
		m.subscriptions = append(m.subscriptions, sensor.Subscription)
		m.pathOutput[sensor.Subscription] = sensor.Output
	}

	m.systemID, _, _ = net.SplitHostPort(m.conn.Target())

	return m
}

// Start gets stream metrics and fan-out to workers.
func (m *MDT) Start(ctx context.Context) error {

	subsArgs := &mdtGRPC.SubscribeRequest{
		RequestId:     int64(os.Getpid()),
		Encode:        gpbkv,
		Subscriptions: m.subscriptions,
		Qos:           &mdtGRPC.QOSMarking{Marking: 10},
	}

	client := mdtGRPC.NewGRPCConfigOperClient(m.conn)
	stream, err := client.CreateSubs(ctx, subsArgs)
	if err != nil {
		return err
	}

	for i := 0; i < 1; i++ {
		go m.worker(ctx)
	}

	for {
		reply, err := stream.Recv()
		// TODO handle error io.EOF
		if err != nil {
			return err
		}

		if reply == nil {
			m.logger.Error("cisco.mdt", zap.String("msg", "nil data"))
			continue
		}

		m.dataChan <- reply.Data
	}

}

func (m *MDT) worker(ctx context.Context) {
	var buf = new(bytes.Buffer)
	for {
		select {
		case d, ok := <-m.dataChan:
			if !ok {
				return
			}

			if err := m.datastore(buf, d); err != nil {
				m.logger.Error("cisco.mdt", zap.Error(err))
			}

		case <-ctx.Done():
			return
		}
	}
}

func (m *MDT) datastore(buf *bytes.Buffer, data []byte) error {
	tm := &mdt.Telemetry{}
	err := proto.Unmarshal(data, tm)
	if err != nil {
		return err
	}

	m.handler(buf, tm)

	return nil
}

func (m *MDT) handler(buf *bytes.Buffer, tm *mdt.Telemetry) {
	var (
		prefix, output string
		timestamp      uint64
		ok             bool
	)

	for _, gpbkv := range tm.DataGpbkv {
		if output, ok = m.pathOutput[tm.GetSubscriptionIdStr()]; !ok {
			m.logger.Error("cisco.mdt", zap.String("msg", "output not found"))
			continue
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
			m.logger.Warn("cisco.mdt", zap.String("msg", "invalid data"))
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
				"system_id": m.systemID,
				"key":       key,
				"value":     value,
			}

			select {
			case m.outChan <- telemetry.ExtDataStore{
				DS:     dataStore,
				Output: output,
			}:
			default:
				m.metrics["dropsTotal"].Inc()
				m.logger.Warn("cisco.mdt", zap.String("error", "dataset drop"))
			}
		}

		buf.Reset()
	}
}

func getKey(buf *bytes.Buffer, kv map[string]interface{}, field *mdt.TelemetryField) {
	// TODO: NX-OS
	if field.Fields != nil {
		if buf.Len() > 0 {
			buf.WriteRune('/')
		}
		buf.WriteString(field.Name)

		for _, subFiled := range field.Fields {
			getKey(buf, kv, subFiled)
		}

	} else {
		kv[buf.String()+"/"+field.Name] = getKeyValue(field)
	}
}

func getTimestamp(t, tm uint64) uint64 {
	if t > 0 {
		return t
	}
	return tm
}

func getKeyLabels(labels map[string]string, field *mdt.TelemetryField) {
	labels[field.Name] = getLabelValue(field)

	for _, subField := range field.Fields {
		getKeyLabels(labels, subField)
	}
}

func getLabelValue(field *mdt.TelemetryField) string {
	switch val := field.ValueByType.(type) {
	case *mdt.TelemetryField_BytesValue:
		return string(val.BytesValue)
	case *mdt.TelemetryField_StringValue:
		return val.StringValue
	case *mdt.TelemetryField_Uint32Value:
		return strconv.FormatUint(uint64(val.Uint32Value), 10)
	case *mdt.TelemetryField_Uint64Value:
		return strconv.FormatUint(val.Uint64Value, 10)
	case *mdt.TelemetryField_Sint32Value:
		return strconv.FormatInt(int64(val.Sint32Value), 10)
	case *mdt.TelemetryField_Sint64Value:
		return strconv.FormatInt(val.Sint64Value, 10)
	case *mdt.TelemetryField_FloatValue:
		return strconv.FormatFloat(float64(val.FloatValue), 'f', -1, 32)
	case *mdt.TelemetryField_DoubleValue:
		return strconv.FormatFloat(val.DoubleValue, 'f', -1, 64)
	case *mdt.TelemetryField_BoolValue:
		if val.BoolValue {
			return "true"
		}
		return "false"

	default:
		return ""
	}
}

func getKeyValue(field *mdt.TelemetryField) interface{} {
	switch val := field.ValueByType.(type) {
	case *mdt.TelemetryField_BytesValue:
		return val.BytesValue
	case *mdt.TelemetryField_BoolValue:
		return val.BoolValue
	case *mdt.TelemetryField_Uint32Value:
		return val.Uint32Value
	case *mdt.TelemetryField_Uint64Value:
		return val.Uint64Value
	case *mdt.TelemetryField_Sint32Value:
		return val.Sint32Value
	case *mdt.TelemetryField_Sint64Value:
		return val.Sint64Value
	case *mdt.TelemetryField_FloatValue:
		return val.FloatValue
	case *mdt.TelemetryField_DoubleValue:
		return val.DoubleValue
	case *mdt.TelemetryField_StringValue:
		if len(val.StringValue) > 0 {
			return val.StringValue
		}
	}

	return nil
}

// Version returns version
func Version() string {
	return mdtVersion
}
