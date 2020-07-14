package consul

import (
	"encoding/json"
	"errors"
	"path"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
)

type consul struct {
	client *api.Client

	prefix    string
	devices   []config.Device
	producers []config.Producer
	databases []config.Database
	global    *config.Global

	informer chan struct{}

	logger *zap.Logger
}

type consulConfig struct {
	Address string
	Prefix  string
}

func New(filename string) (config.Config, error) {
	var (
		err    error
		cfg    = &consulConfig{}
		consul = &consul{informer: make(chan struct{}, 1)}
	)

	if err := yaml.Read(filename, cfg); err != nil {
		return nil, err
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = cfg.Address

	if len(cfg.Prefix) > 0 {
		consul.prefix = cfg.Prefix
	} else {
		consul.prefix = "config/"
	}

	consul.client, err = api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}

	if err = consul.getRemoteConfig(); err != nil {
		return nil, err
	}

	consul.logger = config.GetLogger(consul.global.Logger)

	go consul.watch("keyprefix", consul.prefix, consul.informer)

	return consul, nil
}

func (c *consul) getRemoteConfig() error {
	var (
		err        error
		devicesTpl = make(map[string]config.DeviceTemplate)
		sensors    = make(map[string]*config.Sensor)
	)

	kv := c.client.KV()
	pairs, _, err := kv.List(c.prefix, nil)
	if err != nil {
		return err
	}

	if len(pairs) < 1 {
		return errors.New("consul is empty")
	}

	c.devices = c.devices[:0]
	c.producers = c.producers[:0]
	c.databases = c.databases[:0]

	for _, p := range pairs {
		// skip folder
		if len(p.Value) < 1 {
			continue
		}
		key := strings.TrimPrefix(string(p.Key), c.prefix)
		folder, k := path.Split(key)

		switch folder {
		case "producers/":
			producer := config.Producer{}
			if err := json.Unmarshal(p.Value, &producer); err != nil {
				return err
			}
			producer.Name = k
			c.producers = append(c.producers, producer)
		case "databases/":
			database := config.Database{}
			if err := json.Unmarshal(p.Value, &database); err != nil {
				return err
			}
			database.Name = k
			c.databases = append(c.databases, database)
		case "devices/":
			device := config.DeviceTemplate{}
			if err := json.Unmarshal(p.Value, &device); err != nil {
				return err
			}
			devicesTpl[k] = device
		case "sensors/":
			sensor := config.Sensor{}
			if err := json.Unmarshal(p.Value, &sensor); err != nil {
				return err
			}
			sensors[k] = &sensor
		default:
			if k == "global" {
				err = json.Unmarshal(p.Value, &c.global)
				if err != nil {
					return err
				}
			}
		}
	}

	for _, d := range devicesTpl {
		device := config.ConvDeviceTemplate(d)
		device.Sensors = make(map[string][]*config.Sensor)

		for _, s := range d.Sensors {
			sensor, ok := sensors[s]
			if !ok {
				c.logger.Error("sensor not exist", zap.String("sensor", s))
				continue
			}
			device.Sensors[sensor.Service] = append(device.Sensors[sensor.Service], sensor)
		}

		c.devices = append(c.devices, device)
	}

	// add built-in producer
	c.producers = append(c.producers, config.Producer{
		Name:    "console",
		Service: "console",
	})

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
	return c.getRemoteConfig()
}

func (c *consul) watch(watchType, value string, ch chan<- struct{}) {
	params := make(map[string]interface{})
	params["type"] = watchType

	switch watchType {
	case "keyprefix":
		params["prefix"] = value
	case "key":
		params["key"] = value
	}

	wp, err := watch.Parse(params)
	if err != nil {
		panic(err)
	}

	lastIdx := uint64(0)
	wp.Handler = func(idx uint64, data interface{}) {
		if lastIdx != 0 {
			c.logger.Info("watcher", zap.String("name", value), zap.String("type", watchType))
			select {
			case ch <- struct{}{}:
			default:
			}
		}
		lastIdx = idx
	}

	if err := wp.Run("localhost:8500"); err != nil {
		panic(err)
	}
}
