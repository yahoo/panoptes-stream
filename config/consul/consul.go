package consul

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
)

// consul represents the consul distributed key-value storage
type consul struct {
	client *api.Client

	prefix    string
	devices   []config.Device
	producers []config.Producer
	databases []config.Database
	sensors   []config.Sensor
	global    *config.Global

	informer chan struct{}

	logger *zap.Logger
}

type consulConfig struct {
	Address string
	Prefix  string

	TLSConfig config.TLSConfig
}

// New constructs consul configuration management
func New(filename string) (config.Config, error) {
	var err error

	c := &consul{
		informer: make(chan struct{}, 1),
		global:   &config.Global{},
		logger:   config.GetDefaultLogger(),
	}

	config := &consulConfig{}

	yaml.Read(filename, config)

	prefix := "panoptes_config_consul"
	err = envconfig.Process(prefix, config)
	if err != nil {
		return nil, err
	}

	apiConfig := api.DefaultConfig()
	apiConfig.Address = config.Address

	if config.TLSConfig.Enabled {
		apiConfig.TLSConfig, err = getTLSConfig(config)
		if err != nil {
			return nil, err
		}
	}

	if len(config.Prefix) > 0 {
		c.prefix = config.Prefix
	} else {
		c.prefix = "panoptes/config/"
	}

	c.client, err = api.NewClient(apiConfig)
	if err != nil {
		return nil, err
	}

	if err = c.getRemoteConfig(); err != nil {
		return nil, err
	}

	if !c.global.WatcherDisabled {
		go func() {
			err := c.watch(config.Address)
			if err != nil {
				c.logger.Error("consul.watcher", zap.Error(err))
				os.Exit(1)
			}
		}()
	}

	return c, nil
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
		return fmt.Errorf("consul is empty, prefix=%s", c.prefix)
	}

	c.devices = c.devices[:0]
	c.producers = c.producers[:0]
	c.databases = c.databases[:0]
	c.sensors = c.sensors[:0]

	for _, p := range pairs {
		// skip folder and empty value
		if len(p.Value) < 1 {
			continue
		}
		key := strings.TrimPrefix(string(p.Key), c.prefix)
		folder, k := path.Split(key)

		if !json.Valid(p.Value) {
			return fmt.Errorf("invalid JSON encoding - %s", key)
		}

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

			if err := config.SensorValidation(sensor); err != nil {
				c.logger.Error("consul", zap.Error(err))
				continue
			}
			sensors[k] = &sensor
			c.sensors = append(c.sensors, sensor)
		default:
			if k == "global" {
				err = json.Unmarshal(p.Value, &c.global)
				if err != nil {
					return err
				}

				prefix := "panoptes"
				err = envconfig.Process(prefix, c.global)
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

		if err := config.DeviceValidation(device); err != nil {
			c.logger.Error("consul", zap.Error(err))
			continue
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

// Devices returns configured devices
func (c *consul) Devices() []config.Device {
	return c.devices
}

// Producers returns configured producers
func (c *consul) Producers() []config.Producer {
	return c.producers
}

// Databases returns configured databases
func (c *consul) Databases() []config.Database {
	return c.databases
}

// Sensors returns configured sensors
func (c *consul) Sensors() []config.Sensor {
	return c.sensors
}

// Global returns global configuration
func (c *consul) Global() *config.Global {
	return c.global
}

// Informer returns informer channel
func (c *consul) Informer() chan struct{} {
	return c.informer
}

// Logger returns logging handler
func (c *consul) Logger() *zap.Logger {
	return c.logger
}

// Update gets configuration from consul key value store
func (c *consul) Update() error {
	return c.getRemoteConfig()
}

func (c *consul) watch(addr string) error {
	params := map[string]interface{}{
		"type":   "keyprefix",
		"prefix": c.prefix,
	}

	wp, err := watch.Parse(params)
	if err != nil {
		return err
	}

	lastIdx := uint64(0)
	wp.Handler = func(idx uint64, data interface{}) {
		if lastIdx != 0 {
			c.logger.Info("consul.watcher", zap.String("event", "triggered"), zap.String("prefix", c.prefix))
			select {
			case c.informer <- struct{}{}:
			default:
				c.logger.Warn("consul.watcher", zap.String("event", "notification has been dropped"))
			}
		}
		lastIdx = idx
	}

	if err := wp.Run(addr); err != nil {
		return err
	}

	return nil
}

func getTLSConfig(cfg *consulConfig) (api.TLSConfig, error) {
	var CAPEM []byte

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

		if v, ok := secrets["ca"]; ok {
			CAPEM = v
		}

		return api.TLSConfig{
			CertPEM:            secrets["cert"],
			KeyPEM:             secrets["key"],
			CAPem:              CAPEM,
			InsecureSkipVerify: cfg.TLSConfig.InsecureSkipVerify,
		}, nil

	}

	return api.TLSConfig{
		CertFile:           cfg.TLSConfig.CertFile,
		KeyFile:            cfg.TLSConfig.KeyFile,
		CAFile:             cfg.TLSConfig.CAFile,
		InsecureSkipVerify: cfg.TLSConfig.InsecureSkipVerify,
	}, nil
}
