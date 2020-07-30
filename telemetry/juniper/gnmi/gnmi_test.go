package gnmi

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/generic/gnmi/mock"
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

	cfg := &config.MockConfig{}

	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*500)

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