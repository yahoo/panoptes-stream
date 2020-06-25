package jti

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	jpb "git.vzbuilders.com/marshadrad/panoptes/vendor/juniper/proto/OCJuniper"
	"google.golang.org/grpc"
)

// Register ...
func Register() {
	log.Println("jti registerd")
	telemetry.Register("juniper.jti", New)
}

type KeyValues map[string]interface{}

type JTI struct {
	conn   *grpc.ClientConn
	client jpb.OpenConfigTelemetryClient
	paths  []*jpb.Path

	dataChan chan *jpb.OpenConfigData
}

// New ...
func New(conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.KVChan) telemetry.NMI {
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
			kv := j.decoder(d)
			kv.prettyPrint()
		case <-ctx.Done():
			return
		}
	}
}

func (j *JTI) decoder(d *jpb.OpenConfigData) KeyValues {
	kv := make(KeyValues)
	kv["__service__"] = "jti_v1.0"

	for _, v := range d.Kv {
		switch v.Value.(type) {
		case *jpb.KeyValue_StrValue:
			kv[v.Key] = v.GetStrValue()
			break
		case *jpb.KeyValue_DoubleValue:
			kv[v.Key] = v.GetDoubleValue()
			break
		case *jpb.KeyValue_IntValue:
			kv[v.Key] = v.GetIntValue()
			break
		case *jpb.KeyValue_SintValue:
			kv[v.Key] = v.GetSintValue()
			break
		case *jpb.KeyValue_UintValue:
			kv[v.Key] = v.GetUintValue()
			break
		case *jpb.KeyValue_BytesValue:
			kv[v.Key] = v.GetBytesValue()
			break
		case *jpb.KeyValue_BoolValue:
			kv[v.Key] = v.GetBoolValue()
			break
		}
	}

	jHeaderKV := make(KeyValues)
	jHeaderKV["system_id"] = d.SystemId
	jHeaderKV["component_id"] = d.ComponentId
	jHeaderKV["sub_component_id"] = d.SubComponentId
	jHeaderKV["path"] = d.Path

	jHeaderKV["timestamp"] = d.Timestamp
	jHeaderKV["SequenceNumber"] = d.SequenceNumber

	kv["__juniper_jpb_header__"] = jHeaderKV

	return kv
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

func (kv KeyValues) prettyPrint() error {
	b, err := json.MarshalIndent(kv, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}
