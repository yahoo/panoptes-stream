//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package jti

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/status"
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/juniper/proto/authentication"
	jpb "github.com/yahoo/panoptes-stream/telemetry/juniper/proto/telemetry"
)

var jtiVersion = "1.0"

// JTI represents Junos Telemetry Interface.
type JTI struct {
	conn   *grpc.ClientConn
	client jpb.OpenConfigTelemetryClient
	paths  []*jpb.Path

	dataChan chan *jpb.OpenConfigData
	outChan  telemetry.ExtDSChan
	logger   *zap.Logger

	metrics map[string]status.Metrics

	pathOutput map[string]string
}

// New creates a JTI.
func New(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	var (
		paths      = []*jpb.Path{}
		pathOutput = make(map[string]string)
		metrics    = make(map[string]status.Metrics)
	)

	metrics["gRPCDataTotal"] = status.NewCounter("juniper_jti_grpc_data_total", "")
	metrics["dropsTotal"] = status.NewCounter("juniper_jti_drops_total", "")
	metrics["errorsTotal"] = status.NewCounter("juniper_jti_errors_total", "")
	metrics["processNSecond"] = status.NewGauge("juniper_jti_process_nanosecond", "")

	status.Register(status.Labels{"host": conn.Target()}, metrics)

	for _, sensor := range sensors {
		path := &jpb.Path{
			Path:            sensor.Path,
			SampleFrequency: uint32(sensor.SampleInterval) * 1000,
		}
		paths = append(paths, path)

		if strings.HasSuffix(sensor.Path, "/") {
			pathOutput[sensor.Path] = sensor.Output
		} else {
			pathOutput[fmt.Sprintf("%s/", sensor.Path)] = sensor.Output
		}
	}

	return &JTI{
		logger:     logger,
		conn:       conn,
		paths:      paths,
		dataChan:   make(chan *jpb.OpenConfigData, 100),
		outChan:    outChan,
		pathOutput: pathOutput,
		metrics:    metrics,
	}
}

// Start starts to get stream and fan-out to workers.
func (j *JTI) Start(ctx context.Context) error {
	defer status.Unregister(status.Labels{"host": j.conn.Target()}, j.metrics)

	if err := j.auth(ctx); err != nil {
		return err
	}

	j.client = jpb.NewOpenConfigTelemetryClient(j.conn)
	subClient, err := j.client.TelemetrySubscribe(ctx,
		&jpb.SubscriptionRequest{PathList: j.paths})
	if err != nil {
		return err
	}

	workers := config.GetEnvInt("JUNIPER_JTI_WORKERS", 1)
	for i := 0; i < workers; i++ {
		go j.worker(ctx)
	}

	for {
		resp, err := subClient.Recv()
		if err != nil && ctx.Err() == nil {
			return err
		}

		if ctx.Err() != nil {
			return nil
		}

		j.dataChan <- resp
		j.metrics["gRPCDataTotal"].Inc()
	}

}

func (j *JTI) auth(ctx context.Context) error {
	md, _ := metadata.FromOutgoingContext(ctx)
	if len(md) < 1 {
		return nil
	}

	if len(md["username"]) < 1 || len(md["password"]) < 1 {
		return nil
	}

	rand.Seed(time.Now().UnixNano())
	ID := fmt.Sprintf("panoptes-%d", rand.Intn(1000))

	client := authentication.NewLoginClient(j.conn)
	reply, err := client.LoginCheck(ctx, &authentication.LoginRequest{
		UserName: md["username"][0],
		Password: md["password"][0],
		ClientId: ID,
	})

	if err != nil {
		return err
	}

	if !reply.Result {
		return errors.New("authentiocation failed")
	}

	return err
}

func (j *JTI) worker(ctx context.Context) {
	var (
		regxPath = regexp.MustCompile(`/:(/.*/):`)
		start    time.Time
		rBuf     = new(bytes.Buffer)
		wBuf     = new(bytes.Buffer)
	)

	for {
		select {
		case data, ok := <-j.dataChan:
			if !ok {
				return
			}

			start = time.Now()

			path := regxPath.FindStringSubmatch(data.Path)
			if len(path) < 1 {
				j.metrics["errorsTotal"].Inc()
				j.logger.Error("juniper.jti", zap.String("msg", "path not found"), zap.String("path", data.Path))
				continue
			}
			output, ok := j.pathOutput[path[1]]
			if !ok {
				j.metrics["errorsTotal"].Inc()
				j.logger.Error("juniper.jti", zap.String("msg", "output lookup failed"), zap.String("path", data.Path))
				continue
			}

			j.datastore(rBuf, wBuf, data, output)

			j.metrics["processNSecond"].Set(uint64(time.Since(start).Nanoseconds()))

		case <-ctx.Done():
			return
		}
	}
}

func (j *JTI) datastore(rBuf, wBuf *bytes.Buffer, data *jpb.OpenConfigData, output string) {
	var (
		ds                   telemetry.DataStore
		labels, prefixLabels map[string]string
		prefix               string
	)

	// convert to nanoseconds
	data.Timestamp = data.Timestamp * 1000000

	for _, v := range data.Kv {

		if v.Key == "__prefix__" {
			prefixLabels, prefix = getLabels(rBuf, wBuf, v.GetStrValue())
			if len(prefix) > 1 {
				prefix = prefix[:len(prefix)-1]
			}
			continue
		}

		// skip header
		if prefix == "" {
			continue
		}

		keyLabels, key := getLabels(rBuf, wBuf, v.Key)
		labels = telemetry.MergeLabels(keyLabels, prefixLabels, prefix)

		ds = telemetry.DataStore{
			"prefix":    prefix,
			"labels":    labels,
			"timestamp": data.Timestamp,
			"system_id": data.SystemId,
			"key":       key,
			"value":     getValue(v),
		}

		select {
		case j.outChan <- telemetry.ExtDataStore{
			DS:     ds,
			Output: output,
		}:
		default:
			j.metrics["dropsTotal"].Inc()
			j.logger.Warn("juniper.jti", zap.String("error", "dataset drop"))
		}

	}
}

func getValue(v *jpb.KeyValue) interface{} {
	var value interface{}

	switch v.Value.(type) {
	case *jpb.KeyValue_StrValue:
		value = v.GetStrValue()
	case *jpb.KeyValue_DoubleValue:
		value = v.GetDoubleValue()
	case *jpb.KeyValue_IntValue:
		value = v.GetIntValue()
	case *jpb.KeyValue_SintValue:
		value = v.GetSintValue()
	case *jpb.KeyValue_UintValue:
		value = v.GetUintValue()
	case *jpb.KeyValue_BytesValue:
		value = v.GetBytesValue()
	case *jpb.KeyValue_BoolValue:
		value = v.GetBoolValue()
	}

	return value
}

// Version returns version
func Version() string {
	return jtiVersion
}

func getLabels(r, w *bytes.Buffer, path string) (map[string]string, string) {
	var key, value, s string
	var err error

	r.Reset()
	w.Reset()

	r.WriteString(path)

	labels := make(map[string]string)

	for {
		s, err = r.ReadString('[')
		if err != nil {
			break
		}

		w.WriteString(s[:len(s)-1])

		key, err = r.ReadString('=')
		if err != nil {
			break
		}

		value, err = r.ReadString(']')
		if err != nil {
			break
		}

		if len(key) > 0 && len(value) > 0 {
			if value[0] == '\'' && len(value) > 2 {
				// string
				labels[key[:len(key)-1]] = value[1 : len(value)-2]
			} else {
				// number
				labels[key[:len(key)-1]] = value[:len(value)-1]
			}
		}
	}

	w.WriteString(s)

	return labels, w.String()
}
