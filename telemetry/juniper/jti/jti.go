package jti

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	jpb "git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/proto/OCJuniper"
)

var jtiVersion = "1.0"

// JTI ...
type JTI struct {
	conn   *grpc.ClientConn
	client jpb.OpenConfigTelemetryClient
	paths  []*jpb.Path

	dataChan chan *jpb.OpenConfigData
	outChan  telemetry.ExtDSChan
	lg       *zap.Logger

	pathOutput map[string]string
}

// New ...
func New(lg *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	paths := []*jpb.Path{}
	pathOutput := make(map[string]string)

	for _, sensor := range sensors {
		path := &jpb.Path{
			Path:            sensor.Path,
			SampleFrequency: uint32(sensor.Interval) * 1000,
		}
		paths = append(paths, path)

		if strings.HasSuffix(sensor.Path, "/") {
			pathOutput[sensor.Path] = sensor.Output
		} else {
			pathOutput[fmt.Sprintf("%s/", sensor.Path)] = sensor.Output
		}
	}

	return &JTI{
		conn:       conn,
		paths:      paths,
		dataChan:   make(chan *jpb.OpenConfigData, 100),
		outChan:    outChan,
		pathOutput: pathOutput,
	}
}

// Start ...
func (j *JTI) Start(ctx context.Context) error {
	j.client = jpb.NewOpenConfigTelemetryClient(j.conn)
	stream, err := j.client.TelemetrySubscribe(ctx,
		&jpb.SubscriptionRequest{PathList: j.paths})
	if err != nil {
		return err
	}

	for i := 0; i < 4; i++ {
		go j.worker(ctx)
	}

	for {
		d, err := stream.Recv()
		if err != nil {
			break
		}

		j.dataChan <- d
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
			ds := j.decoder(d)
			path := regxPath.FindStringSubmatch(d.Path)
			if len(path) > 1 {
				output, ok := j.pathOutput[path[1]]
				if !ok {
					j.lg.Warn("path to output not found", zap.String("path", d.Path))
					continue
				}

				select {
				case j.outChan <- telemetry.ExtDataStore{
					DS:     ds,
					Output: output,
				}:
				default:
				}

			} else {
				j.lg.Warn("path not found", zap.String("path", d.Path))
			}

		case <-ctx.Done():
			return
		}
	}
}

func (j *JTI) decoder(d *jpb.OpenConfigData) telemetry.DataStore {
	jHeader := make(telemetry.DataStore)
	jHeader["system_id"] = d.SystemId
	jHeader["component_id"] = d.ComponentId
	jHeader["sub_component_id"] = d.SubComponentId
	jHeader["path"] = d.Path

	jHeader["timestamp"] = d.Timestamp
	jHeader["sequence_number"] = d.SequenceNumber

	jHeader["__service__"] = fmt.Sprintf("jti_v%s", jtiVersion)

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

	return ds
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
