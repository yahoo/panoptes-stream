package jti

import (
	"context"
	"log"
	"regexp"
	"strings"

	jpb "juniper/proto/OCJuniper"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"google.golang.org/grpc"
)

// Register ...
func Register() {
	log.Println("jti registerd")
	telemetry.Register("juniper.jti", New)
}

type JTI struct {
	conn   *grpc.ClientConn
	client jpb.OpenConfigTelemetryClient
	paths  []*jpb.Path

	dataChan chan *jpb.OpenConfigData
}

// New ...
func New(conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	paths := []*jpb.Path{}
	for _, sensor := range sensors {
		path := &jpb.Path{
			Path:            sensor.Path,
			SampleFrequency: uint32(sensor.Interval) * 1000,
		}
		paths = append(paths, path)
	}

	return &JTI{
		conn:     conn,
		paths:    paths,
		dataChan: make(chan *jpb.OpenConfigData, 100),
	}
}

func (j *JTI) subscribe() {
}

// Start ...
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
	}

	return nil
}

func (j *JTI) worker(ctx context.Context) {
	for {
		select {
		case d, ok := <-j.dataChan:
			if !ok {
				return
			}
			ds := j.decoder(d)
			ds.PrettyPrint()
		case <-ctx.Done():
			return
		}
	}
}

func (j *JTI) decoder(d *jpb.OpenConfigData) telemetry.DataStore {
	ds := make(telemetry.DataStore)
	ds["__service__"] = "jti_v1.0"

	for _, v := range d.Kv {
		switch v.Value.(type) {
		case *jpb.KeyValue_StrValue:
			ds[v.Key] = v.GetStrValue()
			break
		case *jpb.KeyValue_DoubleValue:
			ds[v.Key] = v.GetDoubleValue()
			break
		case *jpb.KeyValue_IntValue:
			ds[v.Key] = v.GetIntValue()
			break
		case *jpb.KeyValue_SintValue:
			ds[v.Key] = v.GetSintValue()
			break
		case *jpb.KeyValue_UintValue:
			ds[v.Key] = v.GetUintValue()
			break
		case *jpb.KeyValue_BytesValue:
			ds[v.Key] = v.GetBytesValue()
			break
		case *jpb.KeyValue_BoolValue:
			ds[v.Key] = v.GetBoolValue()
			break
		}
	}

	jHeader := make(telemetry.DataStore)
	jHeader["system_id"] = d.SystemId
	jHeader["component_id"] = d.ComponentId
	jHeader["sub_component_id"] = d.SubComponentId
	jHeader["path"] = d.Path

	jHeader["timestamp"] = d.Timestamp
	jHeader["SequenceNumber"] = d.SequenceNumber

	ds["__juniper_jpb_header__"] = jHeader

	return ds
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
