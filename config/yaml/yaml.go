package yaml

import (
	"io/ioutil"
	"log"

	"git.vzbuilders.com/marshadrad/panoptes/config"

	yml "gopkg.in/yaml.v2"
)

type yaml struct {
	filename  string
	devices   []config.Device
	producers []config.Producer
	global    *config.Global

	informer chan struct{}
}

type producer struct {
	Service    string `yaml:"service"`
	ConfigFile string `yaml:"configFile"`
}

type yamlConfig struct {
	Devices   []config.DeviceTemplate
	Sensors   map[string]config.Sensor
	Producers map[string]producer

	config.Global `yaml:",inline"`
}

// New constructs new yaml config
func New(filename string) (config.Config, error) {
	cfg := &yamlConfig{}
	if err := Read(filename, cfg); err != nil {
		return &yaml{}, err
	}

	y := &yaml{
		filename: filename,

		devices:   configDevices(cfg),
		producers: configProducers(cfg.Producers),
		global:    &cfg.Global,

		informer: make(chan struct{}, 1),
	}

	go y.watcher()

	return y, nil
}

func (y *yaml) Devices() []config.Device {
	return y.devices
}

func (y *yaml) Global() *config.Global {
	return y.global
}

func (y *yaml) Producers() []config.Producer {
	return y.producers
}

func configDevices(y *yamlConfig) []config.Device {
	sensors := make(map[string]*config.Sensor)
	for name, sensor := range y.Sensors {
		sensor := sensor
		sensors[name] = &sensor
	}

	devices := []config.Device{}
	for _, d := range y.Devices {

		device := config.ConvDeviceTemplate(d)
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

func Read(filename string, c interface{}) error {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yml.Unmarshal(b, c)
	if err != nil {
		return err
	}

	return nil
}

func configProducers(p map[string]producer) []config.Producer {
	var producers []config.Producer
	for name, pConfig := range p {
		cfg := make(map[string]interface{})

		if name != "console" {
			if err := Read(pConfig.ConfigFile, &cfg); err != nil {
				log.Fatal(err)
			}
		}

		producers = append(producers, config.Producer{
			Name:    name,
			Service: pConfig.Service,
			Config:  cfg,
		})
	}

	return producers
}
