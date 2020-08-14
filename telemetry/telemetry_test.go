package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

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
		if d.Host == "core1.lax" {
			return false
		}

		return true
	})

	devicesActual = tm.GetDevices()
	assert.Len(t, devicesActual, 1)
	assert.Equal(t, "core1.lhr", devicesActual[0].Host)

	tm.DelFilterOpt("filter1")
	devicesActual = tm.GetDevices()
	assert.Len(t, devicesActual, 2)
	assert.Equal(t, devices, devicesActual)

}
