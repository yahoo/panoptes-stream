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
	"git.vzbuilders.com/marshadrad/panoptes/secret"
)

// Consul represents the Consul distributed key-value storage
type Consul struct {
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

	TLSConfig config.TLSConfig
}

func New(filename string) (config.Config, error) {
	var (
		err    error
		cfg    = &consulConfig{}
		consul = &Consul{informer: make(chan struct{}, 1)}
	)

	if err := yaml.Read(filename, cfg); err != nil {
		return nil, err
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = cfg.Address

	if cfg.TLSConfig.CertFile != "" && !cfg.TLSConfig.Disabled {
		apiConfig.TLSConfig, err = getTLSConfig(cfg)
		if err != nil {
			return nil, err
		}
	}

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

	go consul.watch("keyprefix", consul.prefix, consul.informer)

	return consul, nil
}

func (c *Consul) getRemoteConfig() error {
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
				c.logger = config.GetLogger(c.global.Logger)
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

			if !sensor.Disabled {
				device.Sensors[sensor.Service] = append(device.Sensors[sensor.Service], sensor)
			}
		}

		c.devices = append(c.devices, device)
	}

	// built-in producer
	c.producers = append(c.producers, config.Producer{
		Name:    "console",
		Service: "console",
	})

	return nil
}

func (c *Consul) Devices() []config.Device {
	return c.devices
}

func (c *Consul) Producers() []config.Producer {
	return c.producers
}

func (c *Consul) Databases() []config.Database {
	return c.databases
}

func (c *Consul) Global() *config.Global {
	return c.global
}

func (c *Consul) Informer() chan struct{} {
	return c.informer
}

func (c *Consul) Logger() *zap.Logger {
	return c.logger
}

func (c *Consul) Update() error {
	return c.getRemoteConfig()
}

func (c *Consul) watch(watchType, value string, ch chan<- struct{}) {
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

func getTLSConfig(cfg *consulConfig) (api.TLSConfig, error) {
	sType, path, ok := secret.ParseRemoteSecretInfo(cfg.TLSConfig.CertFile)
	if ok {
		sec, err := secret.GetSecretEngine(sType)
		if err != nil {
			return api.TLSConfig{}, nil
		}

		secrets, err := sec.GetSecrets(path)
		if err != nil {
			return api.TLSConfig{}, nil
		}

		return api.TLSConfig{
			CertPEM: secrets["cert"],
			KeyPEM:  secrets["key"],
		}, nil

	}

	return api.TLSConfig{
		CertFile: cfg.TLSConfig.CertFile,
		KeyFile:  cfg.TLSConfig.KeyFile,
		CAFile:   cfg.TLSConfig.CAFile,
	}, nil
}
