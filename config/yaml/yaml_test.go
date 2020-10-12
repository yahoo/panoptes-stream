//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package yaml

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	yml "gopkg.in/yaml.v3"

	"github.com/yahoo/panoptes-stream/config"
)

var (
	tempDir     string
	yamlContent = `
devices:
  - host: "192.168.59.3"
    password: admin
    port: 50051
    sensors:
      - sensor1
sensors:
  sensor1:
    mode: on_change
    output: "console::stdout"
    path: /interfaces/interface/state
    service: juniper.gnmi
    sampleInterval: 10
    heartbeatInterval: 11
    suppressRedundant: true
producers:
  kafka1:
    service: kafka
    config:
      brokers:
        - 127.0.0.1:9092
      batchSize: 100
      topics:
        - interface
        - bgp
databases:
  influxdb1:
    service: influxdb
    config:
      server: http://localhost:8086
      bucket: mydb
discovery:
  service: pseudo
  config:
    probe: http
    interval: 2
logger:
  level: debug
  encoding: console
status:
  addr: 0.0.0.0:8081
shards:
  enabled: true
deviceOptions:
  username: juniper
`
)

func TestYaml(t *testing.T) {
	tempDir = t.TempDir()

	t.Run("NewYaml", testNewYaml)
	t.Run("Watcher", testWatcher)
	t.Run("NewYamlFileNotExist", testNewYamlFileNotExist)
	t.Run("ConfigDevices", testConfigDevices)
}

func testNewYaml(t *testing.T) {
	cfgFile, err := ioutil.TempFile(tempDir, "config")
	assert.Equal(t, nil, err)

	// write main config
	cfgFile.WriteString(yamlContent)
	cfgFile.Close()

	cfg, err := New(cfgFile.Name())
	assert.Equal(t, nil, err)

	for i := 0; i < 2; i++ {
		// devices
		assert.Len(t, cfg.Devices(), 1)
		assert.Equal(t, "192.168.59.3", cfg.Devices()[0].Host)
		assert.Equal(t, 50051, cfg.Devices()[0].Port)
		assert.Equal(t, "admin", cfg.Devices()[0].Password)
		assert.Contains(t, cfg.Devices()[0].Sensors, "juniper.gnmi")
		assert.Len(t, cfg.Devices()[0].Sensors["juniper.gnmi"], 1)
		assert.Equal(t, "console::stdout", cfg.Devices()[0].Sensors["juniper.gnmi"][0].Output)
		assert.Equal(t, "/interfaces/interface/state", cfg.Devices()[0].Sensors["juniper.gnmi"][0].Path)
		assert.Equal(t, "on_change", cfg.Devices()[0].Sensors["juniper.gnmi"][0].Mode)
		assert.Equal(t, 10, cfg.Devices()[0].Sensors["juniper.gnmi"][0].SampleInterval)
		assert.Equal(t, 11, cfg.Devices()[0].Sensors["juniper.gnmi"][0].HeartbeatInterval)
		assert.Equal(t, true, cfg.Devices()[0].Sensors["juniper.gnmi"][0].SuppressRedundant)

		// producers
		assert.Equal(t, "kafka1", cfg.Producers()[0].Name)
		assert.Equal(t, "kafka", cfg.Producers()[0].Service)
		assert.NotEqual(t, nil, cfg.Producers()[0].Config)

		config := cfg.Producers()[0].Config.(map[string]interface{})
		assert.Equal(t, []interface{}{"127.0.0.1:9092"}, config["brokers"])
		assert.Equal(t, []interface{}{"interface", "bgp"}, config["topics"])

		// databases
		assert.Equal(t, "influxdb1", cfg.Databases()[0].Name)
		assert.Equal(t, "influxdb", cfg.Databases()[0].Service)
		assert.NotEqual(t, nil, cfg.Databases()[0].Config)

		config = cfg.Databases()[0].Config.(map[string]interface{})
		assert.Equal(t, "mydb", config["bucket"])
		assert.Equal(t, "http://localhost:8086", config["server"])

		// Sensors
		sensors := cfg.Sensors()
		assert.Equal(t, "on_change", sensors[0].Mode)
		assert.Equal(t, "juniper.gnmi", sensors[0].Service)

		// global
		assert.Equal(t, "0.0.0.0:8081", cfg.Global().Status.Addr)
		assert.Equal(t, "debug", cfg.Global().Logger["level"])
		assert.Equal(t, true, cfg.Global().Shards.Enabled)
		assert.Equal(t, "juniper", cfg.Global().DeviceOptions.Username)

		// discovery
		assert.Equal(t, "pseudo", cfg.Global().Discovery.Service)
		assert.Equal(t, "http", cfg.Global().Discovery.Config.(map[string]interface{})["probe"])

		// logger
		assert.NotNil(t, cfg.Logger())

		cfg.Update()
	}
}

func testNewYamlFileNotExist(t *testing.T) {
	_, err := New("not_exist.yaml")
	assert.Error(t, err)
}

func testConfigDevices(t *testing.T) {
	y := &yaml{}
	yamlCfg := &yamlConfig{}
	yml.Unmarshal([]byte(yamlContent), yamlCfg)
	devices := y.getDevices(yamlCfg)
	if len(devices) < 1 {
		t.Error("expect to have a device but return ", len(devices))
	}
}

func testWatcher(t *testing.T) {
	cfgFile, err := ioutil.TempFile(tempDir, "configWatcher")
	assert.Equal(t, nil, err)
	defer cfgFile.Close()

	cfgFile.WriteString("foo")

	y := &yaml{
		filename: cfgFile.Name(),
		informer: make(chan struct{}, 1),
		logger:   config.GetDefaultLogger(),
	}

	go func() {
		err := y.watcher()
		assert.Equal(t, nil, err)
	}()

	time.Sleep(time.Second)

	cfgFile.WriteString("bar")

	select {
	case <-y.Informer():
	case <-time.After(time.Second):
		assert.Fail(t, "context deadline exceeded")
	}
}
