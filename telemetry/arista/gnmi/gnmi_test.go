package gnmi

import (
	"bytes"
	"context"
	"testing"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/mock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestAristaSimplePath(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 1)
		ctx     = context.Background()
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.AristaUpdate(), Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := &config.MockConfig{}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "generic.gnmi",
		Output:  "console::stdout",
		Path:    "/interfaces/interface/state/counters",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	resp := <-ch

	assert.Equal(t, sensors[0].Path, resp.DS["prefix"].(string))
	assert.Equal(t, "127.0.0.1", resp.DS["system_id"].(string))
	assert.Equal(t, int64(1595363593437180059), resp.DS["timestamp"].(int64))
	assert.Equal(t, "Ethernet1", resp.DS["labels"].(map[string]string)["name"])
	assert.Equal(t, "out-octets", resp.DS["key"].(string))
	assert.Equal(t, int64(50302030597), resp.DS["value"].(int64))
	assert.Equal(t, "console::stdout", resp.Output)

	assert.Equal(t, "", cfg.LogOutput.String())
}

func TestAristaBGPSimplePath(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 1)
		ctx     = context.Background()
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.AristaBGPUpdate(), Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := &config.MockConfig{}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "generic.gnmi",
		Output:  "console::stdout",
		Path:    "/network-instances/network-instance",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	resp := <-ch

	assert.Equal(t, sensors[0].Path, resp.DS["prefix"].(string))
	assert.Equal(t, "127.0.0.1", resp.DS["system_id"].(string))
	assert.Equal(t, int64(1595363593413814979), resp.DS["timestamp"].(int64))
	assert.Equal(t, "default", resp.DS["labels"].(map[string]string)["name"])
	assert.Equal(t, "BGP", resp.DS["labels"].(map[string]string)["identifier"])
	assert.Equal(t, "IPV6_UNICAST", resp.DS["labels"].(map[string]string)["afi-safi-name"])
	assert.Equal(t, "BGP", resp.DS["labels"].(map[string]string)["/protocols/protocol/name"])
	assert.Equal(t, "protocols/protocol/bgp/global/afi-safis/afi-safi/config/afi-safi-name", resp.DS["key"].(string))
	assert.Equal(t, "openconfig-bgp-types:IPV6_UNICAST", resp.DS["value"].(string))
	assert.Equal(t, "console::stdout", resp.Output)

	assert.Equal(t, "", cfg.LogOutput.String())
}

func TestAristaKVPath(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 1)
		ctx     = context.Background()
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.AristaUpdate(), Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := &config.MockConfig{}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "generic.gnmi",
		Output:  "console::stdout",
		Path:    "/interfaces/interface[name=Ethernet1]/state/counters",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	resp := <-ch

	assert.Equal(t, sensors[0].Path, resp.DS["prefix"].(string))
	assert.Equal(t, "127.0.0.1", resp.DS["system_id"].(string))
	assert.Equal(t, int64(1595363593437180059), resp.DS["timestamp"].(int64))
	assert.Equal(t, "Ethernet1", resp.DS["labels"].(map[string]string)["name"])
	assert.Equal(t, "out-octets", resp.DS["key"].(string))
	assert.Equal(t, int64(50302030597), resp.DS["value"].(int64))
	assert.Equal(t, "console::stdout", resp.Output)

	assert.Equal(t, cfg.LogOutput.String(), "", "unexpected logging")
}

func BenchmarkDS(b *testing.B) {
	var (
		cfg     = &config.MockConfig{}
		buf     = &bytes.Buffer{}
		metrics = make(map[string]status.Metrics)
	)

	metrics["dropsTotal"] = status.NewCounter("generic_gnmi_drops_total", "")

	g := GNMI{
		logger:  cfg.Logger(),
		outChan: make(telemetry.ExtDSChan, 100),
		metrics: metrics,
	}

	g.pathOutput = map[string]string{"/interfaces/interface/state/counters/": "console::stdout"}

	n := mock.AristaUpdate()

	for i := 0; i < b.N; i++ {
		g.datastore(buf, n, n.Update[0], "127.0.0.1")
		<-g.outChan
	}
}
