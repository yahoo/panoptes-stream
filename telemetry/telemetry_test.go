package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"git.vzbuilders.com/marshadrad/panoptes/config"
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

	cfg := &config.MockConfig{MDevices: devices}
	tm := Telemetry{
		cfg:              cfg,
		deviceFilterOpts: DeviceFilterOpts{filterOpts: make(map[string]DeviceFilterOpt)},
	}

	devicesActual := tm.GetDevices()
	assert.Equal(t, devices, devicesActual)

	tm = Telemetry{
		cfg:              cfg,
		deviceFilterOpts: DeviceFilterOpts{filterOpts: make(map[string]DeviceFilterOpt)},
	}

	tm.AddFilterOpt("filter1", func(d config.Device) bool {
		return d.Host != "core1.lax"
	})

	devicesActual = tm.GetDevices()
	assert.Len(t, devicesActual, 1)
	assert.Equal(t, "core1.lhr", devicesActual[0].Host)

	tm.DelFilterOpt("filter1")
	devicesActual = tm.GetDevices()
	assert.Len(t, devicesActual, 2)
	assert.Equal(t, devices, devicesActual)

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
	tm := New(ctx, cfg, telemetryRegistrar, outChan)

	device := config.Device{
		DeviceConfig: config.DeviceConfig{
			Host: "127.0.0.1",
			Port: 50055,
		},
		Sensors: map[string][]*config.Sensor{
			"test.gnmi": {},
		},
	}

	tm.subscribe(device)
	time.Sleep(time.Second * 1)
	assert.Len(t, tm.devices, 1)
	assert.Equal(t, device, tm.devices["127.0.0.1"])
	assert.Contains(t, cfg.LogOutput.String(), "connect")
	assert.Equal(t, uint64(1), tm.metrics["gRPConnCurrent"].Get())

	cancel()
	time.Sleep(time.Second * 1)

	assert.Equal(t, uint64(0), tm.metrics["gRPConnCurrent"].Get())
	assert.Contains(t, cfg.LogOutput.String(), "terminate")
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
