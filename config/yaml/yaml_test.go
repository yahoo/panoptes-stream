package yaml

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"
	"text/template"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"github.com/stretchr/testify/assert"
	yml "gopkg.in/yaml.v3"
)

var yamlContent = `
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
    configFile: {{.KafkaConfigFile}}
databases:
  influxdb1:
    service: influxdb
    configFile: {{.InfluxConfigFile}}		
logger:
  level: debug
  encoding: console
status:
  addr: 0.0.0.0:8081
shard:
  enabled: true
deviceOptions:
  username: juniper       
watcherDisabled: true
`

var kafkaConfig = `
brokers:
  - 127.0.0.1:9092
batchSize: 100
topics:
  - interface
  - bgp
`

var influxConfig = `
server: http://localhost:8086
bucket: mydb
`

type Tpl struct {
	KafkaConfigFile  string
	InfluxConfigFile string
}

func TestNewYaml(t *testing.T) {
	cfgFile, err := ioutil.TempFile("", "config")
	assert.Equal(t, nil, err)
	defer os.Remove(cfgFile.Name())

	cfgKafka, err := ioutil.TempFile("", "kafka")
	assert.Equal(t, nil, err)
	defer os.Remove(cfgKafka.Name())

	cfgInflux, err := ioutil.TempFile("", "influx")
	assert.Equal(t, nil, err)
	defer os.Remove(cfgInflux.Name())

	tpl := Tpl{cfgKafka.Name(), cfgInflux.Name()}
	cfgTpl, err := template.New("test").Parse(yamlContent)
	assert.Equal(t, nil, err)
	buf := new(bytes.Buffer)
	cfgTpl.Execute(buf, tpl)

	// write main config
	cfgFile.WriteString(buf.String())
	cfgFile.Close()

	// write kafka config
	cfgKafka.WriteString(kafkaConfig)
	cfgFile.Close()

	// write influxdb config
	cfgInflux.WriteString(influxConfig)
	cfgInflux.Close()

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

		// global
		assert.Equal(t, "0.0.0.0:8081", cfg.Global().Status.Addr)
		assert.Equal(t, "debug", cfg.Global().Logger["level"])
		assert.Equal(t, true, cfg.Global().Shard.Enabled)
		assert.Equal(t, "juniper", cfg.Global().DeviceOptions.Username)

		cfg.Update()
	}
}

func TestConfigDevices(t *testing.T) {
	y := &yaml{}
	yamlCfg := &yamlConfig{}
	yml.Unmarshal([]byte(yamlContent), yamlCfg)
	devices := y.configDevices(yamlCfg)
	if len(devices) < 1 {
		t.Error("expect to have a device but return ", len(devices))
	}
}

func TestWatcher(t *testing.T) {
	cfgFile, err := ioutil.TempFile("", "config")
	assert.Equal(t, nil, err)
	defer os.Remove(cfgFile.Name())
	cfgFile.WriteString("foo")

	cfg := config.NewMockConfig()

	y := &yaml{
		filename: cfgFile.Name(),
		informer: make(chan struct{}, 1),
		logger:   cfg.Logger(),
	}

	go func() {
		err := y.watcher()
		assert.Equal(t, nil, err)
	}()

	time.Sleep(time.Millisecond * 500)

	cfgFile.WriteString("bar")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	select {
	case <-y.Informer():
	case <-ctx.Done():
		assert.Fail(t, "context deadline exceeded")
	}
}
