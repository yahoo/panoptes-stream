package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var cfg = NewMockConfig()

func TestMock(t *testing.T) {
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

	err := cfg.Update()
	ch := cfg.Informer()

	assert.NoError(t, err)
	assert.Nil(t, ch)
}

func TestUnmarshalSlice(t *testing.T) {
	cfg.LogOutput.Reset()

	cfg.logger.Info("test_info")
	cfg.logger.Error("test_error")

	s := cfg.LogOutput.UnmarshalSlice()
	assert.Len(t, s, 2)
	assert.Equal(t, "test_info", s[0]["M"])
	assert.Equal(t, "test_error", s[1]["M"])

	cfg.LogOutput.Sync()
	cfg.LogOutput.Close()
}
