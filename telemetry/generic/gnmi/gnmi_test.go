package gnmi

import (
	"context"
	"log"
	"testing"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/generic/gnmi/mock"
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
		Path:    "/interfaces/interface/state/counters/",
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
		Path:    "/network-instances/network-instance/",
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
		Path:    "/interfaces/interface[name=Ethernet1]/state/counters/",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	resp := <-ch

	assert.Equal(t, resp.DS["prefix"].(string), sensors[0].Path)
	assert.Equal(t, resp.DS["system_id"].(string), "127.0.0.1")
	assert.Equal(t, resp.DS["timestamp"].(int64), int64(1595363593437180059))
	assert.Equal(t, resp.DS["labels"].(map[string]string)["name"], "Ethernet1")
	assert.Equal(t, resp.DS["key"].(string), "out-octets")
	assert.Equal(t, resp.DS["value"].(int64), int64(50302030597))
	assert.Equal(t, resp.Output, "console::stdout", "unexpected result")

	assert.Equal(t, cfg.LogOutput.String(), "", "unexpected logging")
}

func aTestTwo(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 10)
		ctx     = context.Background()
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.JuniperUpdate(), Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := &config.MockConfig{}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		// TODO
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "generic.gnmi",
		Output:  "console::stdout",
		Path:    "/interfaces/interface/state/counters",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	time.Sleep(1 & time.Second)
	log.Println("log::", cfg.LogOutput.String())
	<-ch

}
