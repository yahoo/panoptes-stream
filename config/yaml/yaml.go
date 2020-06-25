package yaml

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"git.vzbuilders.com/marshadrad/panoptes/config"

	yml "gopkg.in/yaml.v2"
)

type yaml struct {
	devices []config.Device
	global  config.Global
}

type device struct {
	config.DeviceConfig `yaml:",inline"`

	Sensors []string
}

type yamlConfig struct {
	Devices []device
	Sensors map[string]config.Sensor

	config.Global `yaml:",inline"`
}

// LoadConfig constructs new yaml config
func LoadConfig() config.Config {
	cfg := read()

	return &yaml{
		devices: parse(cfg),
		global:  cfg.Global,
	}
}

func (y *yaml) Devices() []config.Device {
	return y.devices
}

func (y *yaml) Global() config.Global {
	return y.global
}

func parse(y *yamlConfig) []config.Device {
	sensors := make(map[string]*config.Sensor)
	for name, sensor := range y.Sensors {
		sensor := sensor
		sensors[name] = &sensor
	}

	devices := []config.Device{}
	for _, d := range y.Devices {

		device := conv(d)
		device.Sensors = make(map[string][]*config.Sensor)

		for _, s := range d.Sensors {
			sensor, ok := sensors[s]
			if !ok {
				log.Fatal("sensor not exist ", s)
			}

			device.Sensors[sensor.Service] = append(device.Sensors[sensor.Service], sensor)
		}

		devices = append(devices, device)
	}

	return devices
}

func read() *yamlConfig {
	cfg := yamlConfig{}
	b, err := ioutil.ReadFile("etc/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	err = yml.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}

func conv(d device) config.Device {
	cd := config.Device{}
	b, _ := json.Marshal(&d)
	json.Unmarshal(b, &cd)
	cd.Sensors = nil
	return cd
}
