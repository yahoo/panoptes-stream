package gnmi

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/mock"
)

func TestGetPrefix(t *testing.T) {
	buf := new(bytes.Buffer)
	g := &GNMI{
		pathOutput: map[string]string{"/interfaces/interface/state/counters/": "out::out"},
	}

	md := mock.CiscoXRInterface()
	prefix, labels, output := g.getPrefix(buf, md.Prefix)

	assert.Equal(t, "/interfaces/interface/state/counters", prefix)
	assert.Equal(t, "out::out", output)
	assert.Equal(t, map[string]string{"name": "GigabitEthernet0/0/0/0"}, labels)
}

func TestGetPrefixWithKey(t *testing.T) {
	buf := new(bytes.Buffer)
	g := &GNMI{
		pathOutput: map[string]string{"/interfaces/interface[name=GigabitEthernet0/0/0/0]/state/counters/": "out::out"},
	}

	md := mock.CiscoXRInterface()
	prefix, labels, output := g.getPrefix(buf, md.Prefix)

	assert.Equal(t, "/interfaces/interface/state/counters", prefix)
	assert.Equal(t, "out::out", output)
	assert.Equal(t, map[string]string{"name": "GigabitEthernet0/0/0/0"}, labels)
}

func TestGetKeys(t *testing.T) {
	buf := new(bytes.Buffer)
	md := mock.CiscoXRInterface()
	key, labels := telemetry.GetKey(buf, md.Update[0].Path.Elem)
	assert.Equal(t, "in-octets", key)
	assert.Equal(t, map[string]string{}, labels)
}

func TestDatastore(t *testing.T) {
	cfg := config.NewMockConfig()
	ch := make(telemetry.ExtDSChan, 20)
	g := GNMI{
		logger:     cfg.Logger(),
		pathOutput: map[string]string{"/interfaces/interface/state/counters/": "out::out"},
		outChan:    ch,
	}

	buf := new(bytes.Buffer)
	md := mock.CiscoXRInterface()
	err := g.datastore(buf, md, "127.0.0.1")
	assert.NoError(t, err)

	for i := 0; i < 12+1; i++ {
		select {
		case m := <-ch:
			assert.Equal(t, int64(1596928627212000000), m.DS["timestamp"])
			assert.Equal(t, map[string]string{"name": "GigabitEthernet0/0/0/0"}, m.DS["labels"])
			assert.Equal(t, "/interfaces/interface/state/counters", m.DS["prefix"])
			assert.Equal(t, "127.0.0.1", m.DS["system_id"])
		default:
			assert.Fail(t, "deadline exceeded")
		}
	}
}

func TestWithMockServer(t *testing.T) {
	var (
		addr    = "127.0.0.1:50555"
		ch      = make(telemetry.ExtDSChan, 20)
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.CiscoXRInterface(), Attempt: 1})
	if err != nil {
		assert.NoError(t, err)
	}
	defer ln.Close()

	cfg := config.NewMockConfig()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		assert.NoError(t, err)
	}

	sensors = append(sensors, &config.Sensor{
		Output: "console::stdout",
		Path:   "/interfaces/interface/state/counters",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)
	for i := 0; i < 12+1; i++ {
		select {
		case <-ch:
		case <-ctx.Done():
			assert.Fail(t, "time deadline exceeded")
		}
	}
}

func TestVersion(t *testing.T) {
	assert.Equal(t, gnmiVersion, Version())
}

func BenchmarkGetPrefix(b *testing.B) {
	buf := new(bytes.Buffer)
	g := &GNMI{
		pathOutput: map[string]string{"/interfaces/interface/state/counters/": "out::out"},
	}

	md := mock.CiscoXRInterface()

	for i := 0; i < b.N; i++ {
		g.getPrefix(buf, md.Prefix)
	}

}

func BenchmarkGetPrefixWithKey(b *testing.B) {
	buf := new(bytes.Buffer)
	g := &GNMI{
		pathOutput: map[string]string{"/interfaces/interface[name=GigabitEthernet0/0/0/0]/state/counters/": "out::out"},
	}

	md := mock.CiscoXRInterface()

	for i := 0; i < b.N; i++ {
		g.getPrefix(buf, md.Prefix)
	}
}
