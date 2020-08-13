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
