package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvDeviceTemplate(t *testing.T) {
	dt := DeviceTemplate{
		DeviceConfig: DeviceConfig{
			Host: "127.0.0.1",
			Port: 50051,
			DeviceOptions: DeviceOptions{
				Username: "user1",
				Password: "pass1",
			},
		},
	}

	d := ConvDeviceTemplate(dt)
	assert.Equal(t, "127.0.0.1", d.Host)
	assert.Equal(t, 50051, d.Port)
	assert.Equal(t, "user1", d.Username)
	assert.Equal(t, "pass1", d.Password)
}

func TestDeviceValidation(t *testing.T) {
	d := Device{}
	assert.NotNil(t, DeviceValidation(d))
	d.Sensors = map[string][]*Sensor{"cisco.gnmi": {{}}}
	assert.NotNil(t, DeviceValidation(d))
	d.Host = "host1"
	assert.NotNil(t, DeviceValidation(d))
	d.Port = 50051
	assert.Nil(t, DeviceValidation(d))

}

func TestGetDefaultLogger(t *testing.T) {
	assert.NotNil(t, GetDefaultLogger())
}

func TestGetLogger(t *testing.T) {
	rawJSON := []byte(`{"level":"info", "encoding": "console", "outputPaths": ["stdout"], "errorOutputPaths":["stderr"]}`)
	lcfg := make(map[string]interface{})
	json.Unmarshal(rawJSON, &lcfg)
	assert.NotNil(t, GetLogger(lcfg))
}
