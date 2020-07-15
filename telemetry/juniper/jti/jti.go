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
	jtiVersion              = "1.0"
	metricTotalReceivedData = status.NewCounter("jti_total_received_data", "")
	labelsRegex             = regexp.MustCompile("(\\/[^\\/]*)\\[([A-Za-z0-9\\-\\/]*\\=[^\\[]*)\\]")
)

func init() {
	status.Register(metricTotalReceivedData)
}

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
	paths := []*jpb.Path{}
	pathOutput := make(map[string]string)

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
		metricTotalReceivedData.Inc()
	}

	return nil
}

func (j *JTI) worker(ctx context.Context) {
	regxPath := regexp.MustCompile("/:(/.*/):")

	for {
		select {
		case d, ok := <-j.dataChan:
			if !ok {
				return
			}
			path := regxPath.FindStringSubmatch(d.Path)
			if len(path) < 1 {
				j.logger.Warn("path not found", zap.String("path", d.Path))
				continue
			}
			output, ok := j.pathOutput[path[1]]
			if !ok {
				j.logger.Warn("path to output not found", zap.String("path", d.Path))
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
		prefix string
		labels = make(map[string]string)
	)

	for _, v := range d.Kv {

		if v.Key == "__prefix__" {
			labels, prefix = getLabels(v.GetStrValue())
			continue
		}

		if prefix == "" {
			continue
		}

		ds = telemetry.DataStore{
			"__prefix__":    prefix,
			"__labels__":    labels,
			"__timestamp__": d.Timestamp * 1000000,
			"__system_id__": d.SystemId,

			v.Key: getValue(v),
		}

		select {
		case j.outChan <- telemetry.ExtDataStore{
			DS:     ds,
			Output: output,
		}:
		default:
			j.logger.Warn("dropped")
		}

	}
}

func (j *JTI) rawDatastore(d *jpb.OpenConfigData, output string) {
	jHeader := telemetry.DataStore{
		"system_id":        d.SystemId,
		"component_id":     d.SystemId,
		"sub_component_id": d.SubComponentId,
		"path":             d.Path,
		"timestamp":        d.Timestamp,
		"sequence_number":  d.SequenceNumber,
		"__service__":      fmt.Sprintf("jti_v%s", jtiVersion),
	}

	dsSlice := []telemetry.DataStore{}
	var ds = make(telemetry.DataStore)

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
	}
}

func getValue(v *jpb.KeyValue) interface{} {
	switch v.Value.(type) {
	case *jpb.KeyValue_StrValue:
		return v.GetStrValue()
	case *jpb.KeyValue_DoubleValue:
		return v.GetDoubleValue()
	case *jpb.KeyValue_IntValue:
		return v.GetIntValue()
	case *jpb.KeyValue_SintValue:
		return v.GetSintValue()
	case *jpb.KeyValue_UintValue:
		return v.GetUintValue()
	case *jpb.KeyValue_BytesValue:
		return v.GetBytesValue()
	case *jpb.KeyValue_BoolValue:
		return v.GetBoolValue()
	}

	return "na"
}

func getJTIPathKValues(p string, valuesOnly bool) []string {
	rgx := regexp.MustCompile("\\/([^\\/]*)\\[([A-Za-z0-9\\-\\/]*\\=[^\\[]*)\\]")
	subs := rgx.FindAllStringSubmatch(p, -1)
	var kv []string

	if len(subs) > 0 {
		for _, sub := range subs {
			if !valuesOnly {
				kv = append(kv, strings.Split(sub[2], "=")[0])
			}

			v := strings.Replace(strings.Split(sub[2], "=")[1], "'", "", -1)
			kv = append(kv, v)
		}
	}
	return kv
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
