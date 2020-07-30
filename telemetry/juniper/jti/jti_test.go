package jti

import (
	"context"
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
		subServer.Send(mock.JuniperLo0InterfaceSample())
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
			value: int64(52613105736),
		},
		{
			key:   "state/counters/in-pkts",
			value: int64(23609955),
		},
		{
			key:   "state/counters/out-octets",
			value: int64(52613105736),
		},
		{
			key:   "state/counters/out-pkts",
			value: int64(23609955),
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

			assert.Equal(t, resp.DS["prefix"].(string), "/interfaces/interface/")
			assert.Equal(t, resp.DS["system_id"].(string), "core1.lax")
			assert.Equal(t, resp.DS["timestamp"].(uint64), uint64(1596067993610)*1000000)
			assert.Equal(t, resp.DS["labels"].(map[string]string)["name"], "lo0")

		case <-ctx.Done():
			assert.Error(t, ctx.Err(), "timeout")
			return
		}
	}

	assert.Equal(t, "", cfg.LogOutput.String())
}
