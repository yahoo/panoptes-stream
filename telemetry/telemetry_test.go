//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/yahoo/panoptes-stream/config"
)

func TestUnsubscribe(t *testing.T) {
	cfg := &config.MockConfig{}
	tm := New(context.Background(), cfg, nil, nil)
	device := config.Device{
		DeviceConfig: config.DeviceConfig{
			Host: "device1",
			Port: 50051,
		},
		Sensors: map[string][]*config.Sensor{},
	}

	// register
	tm.devices["device1"] = device
	_, tm.register["device1"] = context.WithCancel(context.Background())
	tm.metrics["devicesCurrent"].Inc()

	// unregister
	tm.unsubscribe(device)

	assert.Len(t, tm.devices, 0)
	assert.Len(t, tm.register, 0)

	assert.Equal(t, uint64(0), tm.metrics["devicesCurrent"].Get())
}

func TestGetDevices(t *testing.T) {
	devices := []config.Device{
		{
			DeviceConfig: config.DeviceConfig{
				Host: "core1.lax",
			},
		},
		{
			DeviceConfig: config.DeviceConfig{
				Host: "core1.lhr",
			},
		},
	}

	cfg := &config.MockConfig{
		MDevices: devices,
		MGlobal:  &config.Global{},
	}
	tm := Telemetry{
		cfg:              cfg,
		deviceFilterOpts: DeviceFilterOpts{filterOpts: make(map[string]DeviceFilterOpt)},
	}

	devicesActual := tm.GetDevices()
	assert.Len(t, devicesActual, 2)

	cfg.MGlobal.Shards.Enabled = true
	tm.AddFilterOpt("filter1", func(d config.Device) bool {
		return d.Host != "core1.lax"
	})

	devicesActual = tm.GetDevices()
	assert.Len(t, devicesActual, 1)
	assert.Equal(t, "core1.lhr", devicesActual[0].Host)

	tm.DelFilterOpt("filter1")
	devicesActual = tm.GetDevices()
	assert.Len(t, devicesActual, 0)
}

type testGnmi struct{}

func (testGnmi) Start(ctx context.Context) error {
	select {
	case <-time.After(time.Second * 5):
	case <-ctx.Done():
	}
	return nil
}
func testGnmiNew(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan ExtDSChan) NMI {
	return &testGnmi{}
}

func TestSubscribe(t *testing.T) {
	cfg := config.NewMockConfig()
	cfg.MGlobal = &config.Global{}

	outChan := make(ExtDSChan, 100)
	telemetryRegistrar := NewRegistrar(cfg.Logger())
	telemetryRegistrar.Register("test.gnmi", "0.0.0", testGnmiNew)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tm := New(ctx, cfg, telemetryRegistrar, outChan)

	device := config.Device{
		DeviceConfig: config.DeviceConfig{
			Host: "127.0.0.1",
			Port: 50055,
			DeviceOptions: config.DeviceOptions{
				Timeout: 1,
			},
		},
		Sensors: map[string][]*config.Sensor{
			"test.gnmi": {},
		},
	}

	tm.subscribe(device)
	time.Sleep(time.Second * 1)
	assert.Len(t, tm.devices, 1)
	assert.Equal(t, device, tm.devices["127.0.0.1"])
	assert.Equal(t, uint64(0), tm.metrics["gRPConnCurrent"].Get())

	time.Sleep(time.Second * 2)

	assert.Equal(t, uint64(0), tm.metrics["gRPConnCurrent"].Get())
	assert.Contains(t, cfg.LogOutput.String(), "context deadline exceeded")
}

func TestSetCredentials(t *testing.T) {
	cfg := config.NewMockConfig()
	cfg.MGlobal = &config.Global{
		DeviceOptions: config.DeviceOptions{
			Username: "test-g-u",
			Password: "test-g-p",
		},
	}

	tm := &Telemetry{cfg: cfg}
	device := &config.Device{
		DeviceConfig: config.DeviceConfig{
			Host: "127.0.0.1",
			Port: 50055,
		},
		Sensors: map[string][]*config.Sensor{
			"test.gnmi": {},
		},
	}

	ctx, err := tm.setCredentials(context.Background(), device)
	assert.NoError(t, err)
	md, _ := metadata.FromOutgoingContext(ctx)
	assert.Len(t, md, 2)
	assert.Equal(t, "test-g-u", md["username"][0])
	assert.Equal(t, "test-g-p", md["password"][0])

	device.Username = "test-u"
	device.Password = "test-p"
	ctx, _ = tm.setCredentials(context.Background(), device)
	md, _ = metadata.FromOutgoingContext(ctx)
	assert.Len(t, md, 2)
	assert.Equal(t, "test-u", md["username"][0])
	assert.Equal(t, "test-p", md["password"][0])

	device.Username = "__vault::/secrets/path"
	_, err = tm.setCredentials(context.Background(), device)
	assert.Error(t, err)
}

func TestGetTimeout(t *testing.T) {
	cfg := config.NewMockConfig()
	tm := &Telemetry{cfg: cfg}
	to := tm.getTimeout(2)
	assert.Equal(t, 2*time.Second, to)

	to = tm.getTimeout(0)
	assert.Equal(t, 5*time.Second, to)

	cfg.MGlobal.DeviceOptions.Timeout = 4
	to = tm.getTimeout(0)
	assert.Equal(t, 4*time.Second, to)
}
