package consul

import (
	"encoding/json"
	"path"

	"github.com/hashicorp/consul/api"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
)

type consul struct {
	client *api.Client

	filename  string
	devices   []config.Device
	producers []config.Producer
	global    config.Global

	informer chan struct{}
}

type device struct {
	config.DeviceConfig

	Sensors []string
}

type consulConfig struct {
	Address string
}

func New(filename string) (config.Config, error) {
	var (
		err    error
		cfg    = &consulConfig{}
		consul = &consul{}
	)

	if err := yaml.Read(filename, cfg); err != nil {
		return nil, err
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = cfg.Address

	consul.client, err = api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}

	kv := consul.client.KV()

	pairs, _, err := kv.List("config/producers/", nil)
	if err != nil {
		return nil, err
	}

	consul.producers = configProducers(pairs)

	pairs, _, err = kv.List("config/sensors/", nil)
	if err != nil {
		return nil, err
	}

	sensors := configSensors(pairs)

	pairs, _, err = kv.List("config/devices/", nil)
	if err != nil {
		return nil, err
	}

	consul.devices = configDevices(pairs, sensors)

	return consul, nil
}

func (e *consul) Devices() []config.Device {
	return e.devices
}

func (e *consul) Producers() []config.Producer {
	return e.producers
}

func (e *consul) Global() config.Global {
	return e.global
}

func (e *consul) Informer() chan struct{} {

	return e.informer
}

func configProducers(pairs api.KVPairs) []config.Producer {
	var producers []config.Producer

	for _, p := range pairs {
		// skip folder
		if len(p.Value) < 1 {
			continue
		}

		producer := config.Producer{}
		if err := json.Unmarshal(p.Value, &producer); err != nil {
			panic(err)
		}

		_, producer.Name = path.Split(p.Key)
		producers = append(producers, producer)
	}

	return producers
}

func configSensors(pairs api.KVPairs) map[string]*config.Sensor {
	var sensors = make(map[string]*config.Sensor)

	for _, p := range pairs {
		// skip folder
		if len(p.Value) < 1 {
			continue
		}

		sensor := config.Sensor{}
		if err := json.Unmarshal(p.Value, &sensor); err != nil {
			panic(err)
		}

		_, name := path.Split(p.Key)
		sensors[name] = &sensor
	}

	return sensors
}

func configDevices(pairs api.KVPairs, sensors map[string]*config.Sensor) []config.Device {
	devices := []config.Device{}

	for _, p := range pairs {
		// skip folder
		if len(p.Value) < 1 {
			continue
		}

		d := device{}
		if err := json.Unmarshal(p.Value, &d); err != nil {
			panic(err)
		}

		device := conv(d)
		device.Sensors = make(map[string][]*config.Sensor)

		for _, s := range d.Sensors {
			sensor, ok := sensors[s]
			if !ok {
				panic("sensor not exist ", s)
			}

			device.Sensors[sensor.Service] = append(device.Sensors[sensor.Service], sensor)
		}

		devices = append(devices, device)
	}

	return devices
}

func conv(d device) config.Device {
	cd := config.Device{}
	b, _ := json.Marshal(&d)
	json.Unmarshal(b, &cd)
	cd.Sensors = nil
	return cd
}
