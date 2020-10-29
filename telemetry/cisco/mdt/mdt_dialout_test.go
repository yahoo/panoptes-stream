package mdt

import (
	"bytes"
	"context"
	"testing"
	"time"

	dialout "github.com/cisco-ie/nx-telemetry-proto/mdt_dialout"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/mock"
)

func TestDialoutStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := config.NewMockConfig()
	ch := make(telemetry.ExtDSChan, 2)

	cfg.Global().Dialout = config.Dialout{
		Services: map[string]config.DialoutService{
			"cisco.mdt": {
				Addr:    "127.0.0.1:50051",
				Workers: 1,
			},
		},
	}

	d := NewDialout(ctx, cfg, ch)
	go d.Start()
	time.Sleep(time.Second)

	conn, err := grpc.DialContext(ctx, "127.0.0.1:50051", grpc.WithInsecure())
	assert.NoError(t, err)
	mdtDialoutClient := dialout.NewGRPCMdtDialoutClient(conn)
	mdtDialout, err := mdtDialoutClient.MdtDialout(ctx)
	assert.NoError(t, err)

	tm := mock.MDTInterfaceII()
	b, err := proto.Marshal(tm)
	assert.NoError(t, err)
	mdtDialout.Send(&dialout.MdtDialoutArgs{ReqId: 1, Data: b})
	time.Sleep(time.Second)
	r := <-ch
	labels := r.DS["labels"].(map[string]string)
	assert.Equal(t, "Sub3", labels["subscriptionId"])
	assert.Equal(t, "ios", labels["nodeId"])
	assert.Equal(t, "openconfig-interfaces:interfaces/interface", labels["path"])
}

func TestDialoutHandler(t *testing.T) {
	buf := new(bytes.Buffer)
	ch := make(telemetry.ExtDSChan, 10)

	m := &Dialout{
		cfg:        config.NewMockConfig(),
		pathOutput: map[string]string{"Sub3": "console::stdout"},
		outChan:    ch,
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
			assert.Equal(t, "ios", r.DS["system_id"])
			assert.Equal(t, exp.timestamp, r.DS["timestamp"])

		case <-ctx.Done():
			assert.Fail(t, "deadline exceeded")
		}

	}
}

func TestUpdate(t *testing.T) {
	cfg := config.NewMockConfig()
	m := &Dialout{
		cfg:        cfg,
		pathOutput: map[string]string{"Sub3": "console::stdout"},
	}

	cfg.MSensors = []config.Sensor{
		{Subscription: "Sub2", Output: "test"},
	}

	m.Update()

	assert.Equal(t, map[string]string{"Sub2": "test"}, m.pathOutput)
}
