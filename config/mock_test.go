package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMock(t *testing.T) {
	cfg := NewMockConfig()

	assert.Equal(t, "0.0.0", cfg.Global().Version)
	cfg.Logger().Info("hello mock")
	log := cfg.LogOutput.Unmarshal()
	assert.Equal(t, "hello mock", log["M"])
	assert.Equal(t, "INFO", log["L"])

	cfg.MDatabases = []Database{{Name: "influxdb"}}
	cfg.MProducers = []Producer{{Name: "kafka"}}
	cfg.MSensors = []Sensor{{Service: "juniper.gnmi"}}
	cfg.MDevices = []Device{{DeviceConfig: DeviceConfig{Host: "core1.lax"}}}

	assert.Equal(t, []Database{{Name: "influxdb"}}, cfg.Databases())
	assert.Equal(t, []Producer{{Name: "kafka"}}, cfg.Producers())
	assert.Equal(t, []Sensor{{Service: "juniper.gnmi"}}, cfg.Sensors())
	assert.Equal(t, []Device{{DeviceConfig: DeviceConfig{Host: "core1.lax"}}}, cfg.Devices())
}
