//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package mdt

import (
	"bytes"
	"context"
	"testing"
	"time"

	mdtTelemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/mock"
)

func TestHandler(t *testing.T) {
	buf := new(bytes.Buffer)
	ch := make(telemetry.ExtDSChan, 10)

	m := &MDT{
		pathOutput: map[string]string{"Sub3": "console::stdout"},
		outChan:    ch,
		systemID:   "127.0.0.1",
	}

	tm := mock.MDTInterfaceII()
	m.handler(buf, tm)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tt := map[string]struct {
		key       string
		timestamp uint64
		labels    map[string]string
		value     interface{}
	}{
		"GigabitEthernet0/0/0/0state/counters/in-octets": {
			key:       "state/counters/in-octets",
			timestamp: 1597098791076,
			labels:    map[string]string{"name": "GigabitEthernet0/0/0/0", "nodeId": "ios", "path": "openconfig-interfaces:interfaces/interface", "subscriptionId": "Sub3"},
			value:     uint64(1023),
		},
		"GigabitEthernet0/0/0/0state/counters/out-octets": {
			key:       "state/counters/out-octets",
			timestamp: 1597098791076,
			labels:    map[string]string{"name": "GigabitEthernet0/0/0/0", "nodeId": "ios", "path": "openconfig-interfaces:interfaces/interface", "subscriptionId": "Sub3"},
			value:     uint64(872),
		},
		"GigabitEthernet0/0/0/1state/counters/in-octets": {
			key:       "state/counters/in-octets",
			timestamp: 1597098791086,
			labels:    map[string]string{"name": "GigabitEthernet0/0/0/1", "nodeId": "ios", "path": "openconfig-interfaces:interfaces/interface", "subscriptionId": "Sub3"},
			value:     uint64(1223),
		},
		"GigabitEthernet0/0/0/1state/counters/out-octets": {
			key:       "state/counters/out-octets",
			timestamp: 1597098791086,
			labels:    map[string]string{"name": "GigabitEthernet0/0/0/1", "nodeId": "ios", "path": "openconfig-interfaces:interfaces/interface", "subscriptionId": "Sub3"},
			value:     uint64(8172),
		},
	}

	for i := 0; i < 4; i++ {
		select {
		case r := <-ch:
			exp := tt[r.DS["labels"].(map[string]string)["name"]+r.DS["key"].(string)]

			assert.Equal(t, "console::stdout", r.Output)
			assert.Equal(t, exp.value, r.DS["value"])
			assert.Equal(t, exp.labels, r.DS["labels"])
			assert.Equal(t, "127.0.0.1", r.DS["system_id"])
			assert.Equal(t, exp.timestamp, r.DS["timestamp"])

		case <-ctx.Done():
			assert.Fail(t, "deadline exceeded")
		}

	}
}

func TestMDTMockServer(t *testing.T) {
	var (
		addr    = "127.0.0.1:50555"
		ch      = make(telemetry.ExtDSChan, 20)
		sensors []*config.Sensor
	)

	ln, err := mock.StartMDTServer(addr)
	if err != nil {
		assert.NoError(t, err)
	}
	defer ln.Close()

	cfg := config.NewMockConfig()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		assert.NoError(t, err)
	}

	sensors = append(sensors, &config.Sensor{
		Subscription: "Sub3",
		Output:       "test",
	})

	m := New(cfg.Logger(), conn, sensors, ch)
	m.Start(ctx)

	time.Sleep(time.Second)

	for i := 0; i < 4; i++ {
		select {
		case r := <-ch:
			assert.Equal(t, "test", r.Output)
			assert.Equal(t, "127.0.0.1", r.DS["system_id"])
		case <-time.After(time.Second):
			assert.Fail(t, "time exceeded")
		}
	}
}

func TestGetKeyValue(t *testing.T) {
	f := mdtTelemetry.TelemetryField_StringValue{StringValue: "GigabitEthernet0/0/0/0"}
	r := getKeyValue(&mdtTelemetry.TelemetryField{ValueByType: &f})
	assert.Equal(t, "GigabitEthernet0/0/0/0", r)

	u := mdtTelemetry.TelemetryField_Uint32Value{Uint32Value: 5}
	r = getKeyValue(&mdtTelemetry.TelemetryField{ValueByType: &u})
	assert.Equal(t, uint32(5), r)

	b := mdtTelemetry.TelemetryField_BytesValue{BytesValue: []byte("test")}
	r = getKeyValue(&mdtTelemetry.TelemetryField{ValueByType: &b})
	assert.Equal(t, []byte("test"), r)

	bo := mdtTelemetry.TelemetryField_BoolValue{BoolValue: true}
	r = getKeyValue(&mdtTelemetry.TelemetryField{ValueByType: &bo})
	assert.Equal(t, true, r)

	s := mdtTelemetry.TelemetryField_Sint32Value{Sint32Value: 5}
	r = getKeyValue(&mdtTelemetry.TelemetryField{ValueByType: &s})
	assert.Equal(t, int32(5), r)

	ss := mdtTelemetry.TelemetryField_Sint64Value{Sint64Value: 5}
	r = getKeyValue(&mdtTelemetry.TelemetryField{ValueByType: &ss})
	assert.Equal(t, int64(5), r)

	d := mdtTelemetry.TelemetryField_DoubleValue{DoubleValue: 5}
	r = getKeyValue(&mdtTelemetry.TelemetryField{ValueByType: &d})
	assert.Equal(t, float64(5), r)

	ff := mdtTelemetry.TelemetryField_FloatValue{FloatValue: 5.5}
	r = getKeyValue(&mdtTelemetry.TelemetryField{ValueByType: &ff})
	assert.Equal(t, float32(5.5), r)
}
