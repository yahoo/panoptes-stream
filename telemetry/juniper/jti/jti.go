package jti

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	jpb "git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/proto/OCJuniper"
)

var (
	jtiVersion = "1.0"

	labelsRegex = regexp.MustCompile(`(\/[^\/]*)\[([A-Za-z0-9\-\/]*\=[^\[]*)\]`)

	metricGRPCDataTotal = status.NewCounter("juniper_jti_grpc_data_total", "")
	metricJTIDropsTotal = status.NewCounter("juniper_jti_drops_total", "")
	metricErrorsTotal   = status.NewCounter("juniper_jti_errors_total", "")
)

// JTI represents Junos Telemetry Interface.
type JTI struct {
	conn   *grpc.ClientConn
	client jpb.OpenConfigTelemetryClient
	paths  []*jpb.Path

	dataChan chan *jpb.OpenConfigData
	outChan  telemetry.ExtDSChan
	logger   *zap.Logger

	pathOutput map[string]string
}

// New creates a JTI.
func New(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	var (
		paths      = []*jpb.Path{}
		pathOutput = make(map[string]string)
	)

	status.Register(
		status.Labels{"host": conn.Target()},
		metricGRPCDataTotal,
		metricJTIDropsTotal,
		metricErrorsTotal,
	)

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
	}
}

// Start starts to get stream and fan-out to workers.
func (j *JTI) Start(ctx context.Context) error {
	j.client = jpb.NewOpenConfigTelemetryClient(j.conn)
	stream, err := j.client.TelemetrySubscribe(ctx,
		&jpb.SubscriptionRequest{PathList: j.paths})
	if err != nil {
		return err
	}

	for i := 0; i < 1; i++ {
		go j.worker(ctx)
	}

	for {
		d, err := stream.Recv()
		if err != nil {
			break
		}

		j.dataChan <- d
		metricGRPCDataTotal.Inc()
	}

	return nil
}

func (j *JTI) worker(ctx context.Context) {
	var regxPath = regexp.MustCompile(`/:(/.*/):`)

	for {
		select {
		case d, ok := <-j.dataChan:
			if !ok {
				return
			}
			path := regxPath.FindStringSubmatch(d.Path)
			if len(path) < 1 {
				metricErrorsTotal.Inc()
				j.logger.Warn("juniper.jti", zap.String("msg", "path not found"), zap.String("path", d.Path))
				continue
			}
			output, ok := j.pathOutput[path[1]]
			if !ok {
				metricErrorsTotal.Inc()
				j.logger.Warn("juniper.jti", zap.String("msg", "output lookup failed"), zap.String("path", d.Path))
				continue
			}

			if isRawRequested(output) {
				j.rawDatastore(d, output)
			} else {
				j.datastore(d, output)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (j *JTI) datastore(d *jpb.OpenConfigData, output string) {
	var (
		ds     = make(telemetry.DataStore)
		labels = make(map[string]string)
		prefix string
	)

	for _, v := range d.Kv {

		if v.Key == "__prefix__" {
			labels, prefix = getLabels(v.GetStrValue())
			continue
		}

		// skip header
		if prefix == "" {
			continue
		}

		ds = telemetry.DataStore{
			"__prefix":    prefix,
			"__labels":    labels,
			"__timestamp": d.Timestamp * 1000000,
			"__system_id": d.SystemId,

			v.Key: getValue(v),
		}

		select {
		case j.outChan <- telemetry.ExtDataStore{
			DS:     ds,
			Output: output,
		}:
		default:
			metricJTIDropsTotal.Inc()
			j.logger.Warn("juniper.jti", zap.String("error", "dataset drop"))
		}

	}
}

func (j *JTI) rawDatastore(d *jpb.OpenConfigData, output string) {
	var (
		ds      = make(telemetry.DataStore)
		dsSlice = []telemetry.DataStore{}
		jHeader = telemetry.DataStore{
			"system_id":        d.SystemId,
			"component_id":     d.SystemId,
			"sub_component_id": d.SubComponentId,
			"path":             d.Path,
			"timestamp":        d.Timestamp,
			"sequence_number":  d.SequenceNumber,
			"__service__":      fmt.Sprintf("jti_v%s", jtiVersion),
		}
	)

	for _, v := range d.Kv {

		if v.Key == "__prefix__" {
			if _, ok := ds[v.Key]; ok {
				dsSlice = append(dsSlice, ds)
				ds = make(telemetry.DataStore)
			}

			ds[v.Key] = v.GetStrValue()
			continue
		}

		if _, ok := ds["__prefix__"]; !ok {
			jHeader[v.Key] = getValue(v)
			continue
		}

		ds[v.Key] = getValue(v)

	}

	// last dataset
	dsSlice = append(dsSlice, ds)
	ds = make(telemetry.DataStore)
	ds["__juniper_telemetry_header__"] = jHeader
	ds["dataset"] = dsSlice

	select {
	case j.outChan <- telemetry.ExtDataStore{
		DS:     ds,
		Output: output,
	}:
	default:
		metricJTIDropsTotal.Inc()
		j.logger.Warn("juniper.jti", zap.String("error", "dataset drop"))
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

func Version() string {
	return jtiVersion
}

func isRawRequested(output string) bool {
	return strings.HasSuffix(output, "::raw")
}

func getLabels(prefix string) (map[string]string, string) {
	labels := make(map[string]string)
	subs := labelsRegex.FindAllStringSubmatch(prefix, -1)
	for _, sub := range subs {
		if len(sub) != 3 {
			continue
		}
		kv := strings.Split(sub[2], "=")
		labels[kv[0]] = strings.ReplaceAll(kv[1], "'", "")
		prefix = strings.Replace(prefix, sub[0], sub[1], 1)
	}

	return labels, prefix
}
