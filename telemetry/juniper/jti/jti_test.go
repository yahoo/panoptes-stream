package jti

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/jti/mock"
	jpb "git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/proto/OCJuniper"
)

type Update struct{}

func (u *Update) Run(subReq *jpb.SubscriptionRequest, subServer jpb.OpenConfigTelemetry_TelemetrySubscribeServer) error {
	switch subReq.PathList[0].Path {
	case "/interfaces/interface[name='lo0']/state/counters/":
		return subServer.Send(mock.JuniperLo0InterfaceSample())
	case "/network-instances/network-instance/protocols/protocol/bgp/":
		return subServer.Send(mock.JuniperBGPSample())
	}
	return nil
}

func TestWithJTIServer(t *testing.T) {
	var addr = "127.0.0.1:50500"

	ln, err := mock.StartJTIServer(addr, &Update{})
	if err != nil {

	}
	defer ln.Close()

	t.Run("JuniperLo0InterfaceSample", JuniperLo0InterfaceSample)
	t.Run("JuniperBGPSample", JuniperBGPSample)
}

func JuniperLo0InterfaceSample(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 10)
		sensors []*config.Sensor
	)

	cfg := &config.MockConfig{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "juniper.jti",
		Output:  "console::stdout",
		Path:    "/interfaces/interface[name='lo0']/state/counters/",
	})

	j := New(cfg.Logger(), conn, sensors, ch)
	j.Start(ctx)

	expected := []struct {
		key   string
		value interface{}
	}{
		{
			key:   "state/counters/in-octets",
			value: uint64(52613105736),
		},
		{
			key:   "state/counters/in-pkts",
			value: uint64(23609955),
		},
		{
			key:   "state/counters/out-octets",
			value: uint64(52613105736),
		},
		{
			key:   "state/counters/out-pkts",
			value: uint64(23609955),
		},
		{
			key:   "state/counters/last-clear",
			value: "Never",
		},
	}

	for i := 0; i < 5; i++ {
		select {
		case resp := <-ch:
			assert.Equal(t, expected[i].key, resp.DS["key"].(string))
			assert.Equal(t, expected[i].value, resp.DS["value"])

			assert.Equal(t, "/interfaces/interface/", resp.DS["prefix"].(string))
			assert.Equal(t, "core1.lax", resp.DS["system_id"].(string))
			assert.Equal(t, uint64(1596067993610)*1000000, resp.DS["timestamp"].(uint64))
			assert.Equal(t, "lo0", resp.DS["labels"].(map[string]string)["name"])

		case <-ctx.Done():
			assert.Fail(t, "context deadline exceeded")
			return
		}
	}

	assert.Equal(t, "", cfg.LogOutput.String())
}

func JuniperBGPSample(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 100)
		sensors []*config.Sensor
		prefix  string
		labels  map[string]string
	)

	cfg := &config.MockConfig{}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "juniper.jti",
		Output:  "console::stdout",
		Path:    "/network-instances/network-instance/protocols/protocol/bgp/",
	})

	j := New(cfg.Logger(), conn, sensors, ch)
	j.Start(ctx)

	KV := mock.JuniperBGPSample().Kv

	for _, metric := range KV {
		if metric.Key == "__prefix__" {
			labels, prefix = getLabels(getValue(metric).(string))
			continue
		}

		if strings.HasPrefix(metric.Key, "__") {
			continue
		}

		select {
		case resp := <-ch:
			assert.Equal(t, labels, resp.DS["labels"].(map[string]string))
			assert.Equal(t, prefix, resp.DS["prefix"])
			assert.Equal(t, metric.Key, resp.DS["key"])
			assert.Equal(t, getValue(metric), resp.DS["value"])
			assert.Equal(t, "core1.lax", resp.DS["system_id"])
			assert.Equal(t, uint64(1596087032354*1000000), resp.DS["timestamp"].(uint64))

		case <-ctx.Done():
			assert.Fail(t, "context deadline exceeded")
			return
		}
	}
}
