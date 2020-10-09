//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package juniper

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/telemetry"
	jGNMI "github.com/yahoo/panoptes-stream/telemetry/juniper/gnmi"
	"github.com/yahoo/panoptes-stream/telemetry/mock"
	"google.golang.org/grpc"
)

func TestJuniper(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	gServer := grpc.NewServer()
	juniperGnmiUpdate := New(ctx, 5)
	mockServer := &mock.GNMIServer{Resp: juniperGnmiUpdate}
	gnmi.RegisterGNMIServer(gServer, mockServer)
	go gServer.Serve(ln)

	cfg := config.NewMockConfig()
	ch := make(telemetry.ExtDSChan, 500)
	sensors := []*config.Sensor{
		{Service: "juniper.gnmi",
			Mode:   "sample",
			Path:   "/interfaces/interface/state/counters/",
			Output: "test::test"},
	}
	conn, err := grpc.DialContext(ctx, ln.Addr().String(), grpc.WithInsecure())
	assert.NoError(t, err)
	g := jGNMI.New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	t.Log(cfg.LogOutput.String())

	counter := 0
	metrics := make(map[string]int)
L:
	for {
		select {
		case metric := <-ch:
			counter++
			metrics[metric.DS["key"].(string)]++
		case <-time.After(time.Second):
			break L
		}
	}

	assert.Equal(t, 240, counter)

	assert.Equal(t, 48, metrics["state/counters/in-octets"])
	assert.Equal(t, 48, metrics["state/counters/in-pkts"])
	assert.Equal(t, 48, metrics["state/counters/last-clear"])
	assert.Equal(t, 48, metrics["state/counters/out-octets"])
	assert.Equal(t, 48, metrics["state/counters/out-pkts"])
}
