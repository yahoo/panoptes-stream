package consul

import (
	"encoding/json"
	"fmt"
	"path"

	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
)

type consul struct {
	client *api.Client

	filename  string
	devices   []config.Device
	producers []config.Producer
	databases []config.Database
	global    *config.Global

	informer chan struct{}

	logger *zap.Logger
}

type consulConfig struct {
	Address string
}

func New(filename string) (config.Config, error) {
	var (
		err    error
		cfg    = &consulConfig{}
		consul = &consul{
			informer: make(chan struct{}, 1),
		}
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

	err = consul.getKVConfig()
	if err != nil {
		return nil, err
	}

	consul.logger = config.GetLogger(consul.global.Logger)

	go consul.watch("keyprefix", "config/", nil)

	return consul, nil
}

func (c *consul) getKVConfig() error {
	var err error

	kv := c.client.KV()

	c.producers, err = configProducers(kv, "config/producers/")
	if err != nil {
		return err
	}

	c.databases, err = configDatabases(kv, "config/databases/")
	if err != nil {
		return err
	}

	sensors, err := configSensors(kv, "config/sensors/")
	if err != nil {
		return err
	}

	c.devices, err = configDevices(kv, "config/devices/", sensors)
	if err != nil {
		return err
	}

	c.global, err = configdGlobal(kv, "config/global")
	if err != nil {
		return err
	}

	return nil
}

func (c *consul) Devices() []config.Device {
	return c.devices
}

func (c *consul) Producers() []config.Producer {
	return c.producers
}

func (c *consul) Databases() []config.Database {
	return c.databases
}

func (c *consul) Global() *config.Global {
	return c.global
}

func (c *consul) Informer() chan struct{} {

	return c.informer
}

func (c *consul) Logger() *zap.Logger {
	return c.logger
}

func (c *consul) Update() error {
	return c.getKVConfig()
}

func configProducers(kv *api.KV, prefix string) ([]config.Producer, error) {
	var producers []config.Producer

	pairs, _, err := kv.List(prefix, nil)
	if err != nil {
		return nil, err
	}

	for _, p := range pairs {
		// skip folder
		if len(p.Value) < 1 {
			continue
		}

		producer := config.Producer{}
		if err := json.Unmarshal(p.Value, &producer); err != nil {
			return nil, err
		}

		_, producer.Name = path.Split(p.Key)
		producers = append(producers, producer)
	}

	return producers, nil
}

func configDatabases(kv *api.KV, prefix string) ([]config.Database, error) {
	var databases []config.Database

	pairs, _, err := kv.List(prefix, nil)
	if err != nil {
		return nil, err
	}

	for _, p := range pairs {
		// skip folder
		if len(p.Value) < 1 {
			continue
		}

		database := config.Database{}
		if err := json.Unmarshal(p.Value, &database); err != nil {
			return nil, err
		}

		_, database.Name = path.Split(p.Key)
		databases = append(databases, database)
	}

	return databases, nil
}

func configSensors(kv *api.KV, prefix string) (map[string]*config.Sensor, error) {
	var sensors = make(map[string]*config.Sensor)

	pairs, _, err := kv.List(prefix, nil)
	if err != nil {
		return nil, err
	}

	for _, p := range pairs {
		// skip folder
		if len(p.Value) < 1 {
			continue
		}

		sensor := config.Sensor{}
		if err := json.Unmarshal(p.Value, &sensor); err != nil {
			return nil, err
		}

		_, name := path.Split(p.Key)
		sensors[name] = &sensor
	}

	return sensors, nil
}

func configDevices(kv *api.KV, prefix string, sensors map[string]*config.Sensor) ([]config.Device, error) {
	devices := []config.Device{}

	pairs, _, err := kv.List(prefix, nil)
	if err != nil {
		return nil, err
	}

	for _, p := range pairs {
		// skip folder
		if len(p.Value) < 1 {
			continue
		}

		d := config.DeviceTemplate{}
		if err := json.Unmarshal(p.Value, &d); err != nil {
			panic(err)
		}

		device := config.ConvDeviceTemplate(d)
		device.Sensors = make(map[string][]*config.Sensor)

		for _, s := range d.Sensors {
			sensor, ok := sensors[s]
			if !ok {
				return nil, fmt.Errorf("%s sensor not exist", s)
			}

			device.Sensors[sensor.Service] = append(device.Sensors[sensor.Service], sensor)
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func configdGlobal(kv *api.KV, prefix string) (*config.Global, error) {
	global := &config.Global{}

	pair, _, err := kv.Get(prefix, nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(pair.Value, global)
	if err != nil {
		return nil, err
	}

	return global, nil
}
