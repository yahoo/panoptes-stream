//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package gnmi

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/status"
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/mock"
)

func TestJuniperCountersMock(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 10)
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.JuniperUpdate(), Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := config.NewMockConfig()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "juniper.gnmi",
		Output:  "console::stdout",
		Path:    "/interfaces/interface/state/counters",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	expected := []struct {
		key   string
		value interface{}
	}{
		{
			key:   "state/counters/in-pkts",
			value: int64(23004050),
		},
		{
			key:   "state/counters/out-pkts",
			value: int64(23004056),
		},
		{
			key:   "state/counters/in-octets",
			value: int64(50302030597),
		},
		{
			key:   "state/counters/out-octets",
			value: int64(50302030492),
		},
		{
			key:   "state/counters/last-clear",
			value: "Never",
		},
	}

	r := make(map[string]telemetry.ExtDataStore)

	for i := 0; i < 5; i++ {
		select {
		case resp := <-ch:
			r[resp.DS["key"].(string)] = resp

			assert.Equal(t, "/interfaces/interface", resp.DS["prefix"].(string))
			assert.Equal(t, "127.0.0.1", resp.DS["system_id"].(string))
			assert.Equal(t, int64(1595951912880990837), resp.DS["timestamp"].(int64))
			assert.Equal(t, "lo0", resp.DS["labels"].(map[string]string)["name"])

		case <-ctx.Done():
			assert.Fail(t, "context deadline exceeded")
			return
		}
	}

	for _, e := range expected {
		resp := r[e.key]
		assert.Equal(t, e.value, resp.DS["value"])
	}

	assert.Equal(t, "", cfg.LogOutput.String())
}

func TestJuniperKeyLabel(t *testing.T) {
	var (
		cfg     = config.NewMockConfig()
		buf     = &bytes.Buffer{}
		metrics = make(map[string]status.Metrics)
		ch      = make(telemetry.ExtDSChan, 1)
	)

	metrics["dropsTotal"] = status.NewCounter("juniper_gnmi_drops_total", "")

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	g := &GNMI{
		logger:     cfg.Logger(),
		outChan:    ch,
		metrics:    metrics,
		pathOutput: map[string]string{"/interfaces/interface/state/counters/": "console::stdout"},
	}

	g.datastore(buf, &gnmi.SubscribeResponse_Update{Update: mock.JuniperFakeKeyLabel()}, "127.0.0.1")

	select {
	case resp := <-ch:
		assert.Equal(t, map[string]string{"name": "lo0", "queue-number": "2"}, resp.DS["labels"])
	case <-ctx.Done():
		assert.Fail(t, "context deadline exceeded")
	}
}

func TestJunipeDuplicateLabel(t *testing.T) {
	var (
		cfg     = config.NewMockConfig()
		buf     = &bytes.Buffer{}
		metrics = make(map[string]status.Metrics)
		ch      = make(telemetry.ExtDSChan, 1)
	)

	metrics["dropsTotal"] = status.NewCounter("juniper_gnmi_drops_total", "")

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	g := &GNMI{
		logger:     cfg.Logger(),
		outChan:    ch,
		metrics:    metrics,
		pathOutput: map[string]string{"/interfaces/interface/state/counters/": "console::stdout"},
	}

	g.datastore(buf, &gnmi.SubscribeResponse_Update{Update: mock.JuniperFakeDuplicateLabel()}, "127.0.0.1")

	select {
	case resp := <-ch:
		assert.Equal(t, map[string]string{"/interfaces/interface/name": "lo0", "name": "fake"}, resp.DS["labels"])
	case <-ctx.Done():
		assert.Fail(t, "context deadline exceeded")
	}
}

func BenchmarkDS(b *testing.B) {
	cfg := config.NewMockConfig()
	buf := &bytes.Buffer{}
	metrics := make(map[string]status.Metrics)

	metrics["dropsTotal"] = status.NewCounter("juniper_gnmi_drops_total", "")

	g := &GNMI{
		logger:  cfg.Logger(),
		outChan: make(telemetry.ExtDSChan, 100),
		metrics: metrics,
	}

	g.pathOutput = map[string]string{"/interfaces/interface/state/counters/": "console::stdout"}

	update := &gnmi.SubscribeResponse_Update{
		Update: mock.JuniperUpdate(),
	}

	for i := 0; i < b.N; i++ {
		g.datastore(buf, update, "core1.lax")
		buf.Reset()
		<-g.outChan
	}
}
