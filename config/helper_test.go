//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

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
	assert.NotNil(t, GetLogger(nil))
}

func TestSensorValidation(t *testing.T) {
	sensor := Sensor{Service: "juniper.gnmi"}
	err := SensorValidation(sensor)
	assert.NoError(t, err)
	sensor = Sensor{Service: "noname.gnmi"}
	err = SensorValidation(sensor)
	assert.Error(t, err)
}

func TestSetDefaultGlobal(t *testing.T) {
	g := Global{}
	SetDefaultGlobal(&g)
	assert.Greater(t, g.BufferSize, 0)
	assert.Greater(t, g.OutputBufferSize, 0)
	assert.NotEmpty(t, g.Version)
}

func TestSetDefault(t *testing.T) {
	v := 0
	SetDefault(&v, 5)
	assert.Equal(t, 5, v)

	SetDefault(&v, 6)
	assert.Equal(t, 5, v)

	vv := uint(0)
	SetDefault(&vv, 5)
	assert.Equal(t, uint(5), vv)

	SetDefault(&vv, 6)
	assert.Equal(t, uint(5), vv)
}
