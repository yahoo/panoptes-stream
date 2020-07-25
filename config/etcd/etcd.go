package etcd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"path"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
)

type etcd struct {
	client *clientv3.Client

	prefix    string
	devices   []config.Device
	producers []config.Producer
	databases []config.Database
	global    *config.Global

	informer chan struct{}

	logger *zap.Logger
}

type etcdConfig struct {
	Endpoints []string
	Prefix    string

	TLSConfig config.TLSConfig
}

// New creates an etcd configuration
func New(filename string) (config.Config, error) {
	var (
		err       error
		tlsConfig *tls.Config
		config    = &etcdConfig{}
		etcd      = &etcd{informer: make(chan struct{}, 1)}
	)

	if err := yaml.Read(filename, config); err != nil {
		return nil, err
	}

	prefix := "panoptes_config_etcd"
	err = envconfig.Process(prefix, config)
	if err != nil {
		return nil, err
	}

	if len(config.Prefix) > 0 {
		etcd.prefix = config.Prefix
	} else {
		etcd.prefix = "config/"
	}

	if config.TLSConfig.CertFile != "" && !config.TLSConfig.Disabled {
		tlsConfig, err = secret.GetTLSConfig(&config.TLSConfig)
		if err != nil {
			return nil, err
		}
	}

	etcd.client, err = clientv3.New(clientv3.Config{
		Endpoints: config.Endpoints,
		TLS:       tlsConfig,
	})
	if err != nil {
		return nil, err
	}

	if err = etcd.getRemoteConfig(); err != nil {
		return nil, err
	}

	go etcd.watch(etcd.informer)

	return etcd, nil
}

func (e *etcd) getRemoteConfig() error {
	var (
		devicesTpl = make(map[string]config.DeviceTemplate)
		sensors    = make(map[string]*config.Sensor)
	)

	requestTimeout, _ := time.ParseDuration("5s")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	resp, err := e.client.Get(ctx, e.prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return err
	}

	e.devices = e.devices[:0]
	e.producers = e.producers[:0]
	e.databases = e.databases[:0]

	if len(resp.Kvs) < 1 {
		return errors.New("etcd is empty")
	}

	for _, ev := range resp.Kvs {
		key := strings.TrimPrefix(string(ev.Key), e.prefix)
		folder, k := path.Split(key)

		switch folder {
		case "producers/":
			producer := config.Producer{}
			if err := json.Unmarshal(ev.Value, &producer); err != nil {
				return err
			}
			producer.Name = k
			e.producers = append(e.producers, producer)
		case "databases/":
			database := config.Database{}
			if err := json.Unmarshal(ev.Value, &database); err != nil {
				return err
			}
			database.Name = k
			e.databases = append(e.databases, database)
		case "devices/":
			device := config.DeviceTemplate{}
			if err := json.Unmarshal(ev.Value, &device); err != nil {
				return err
			}
			devicesTpl[k] = device
		case "sensors/":
			sensor := config.Sensor{}
			if err := json.Unmarshal(ev.Value, &sensor); err != nil {
				return err
			}
			sensors[k] = &sensor
		default:
			if k == "global" {
				err = json.Unmarshal(ev.Value, &e.global)
				if err != nil {
					return err
				}
				e.logger = config.GetLogger(e.global.Logger)
			}
		}
	}

	for _, d := range devicesTpl {
		device := config.ConvDeviceTemplate(d)
		device.Sensors = make(map[string][]*config.Sensor)

		for _, s := range d.Sensors {
			sensor, ok := sensors[s]
			if !ok {
				e.logger.Error("sensor not exist", zap.String("sensor", s))
				continue
			}

			if !sensor.Disabled {
				device.Sensors[sensor.Service] = append(device.Sensors[sensor.Service], sensor)
			}
		}

		e.devices = append(e.devices, device)
	}

	// built-in producer
	e.producers = append(e.producers, config.Producer{
		Name:    "console",
		Service: "console",
	})

	return nil
}

func (e *etcd) Devices() []config.Device {
	return e.devices
}

func (e *etcd) Producers() []config.Producer {
	return e.producers
}

func (e *etcd) Databases() []config.Database {
	return e.databases
}

func (e *etcd) Global() *config.Global {
	return e.global
}

func (e *etcd) Informer() chan struct{} {
	return e.informer
}

func (e *etcd) Logger() *zap.Logger {
	return e.logger
}

func (e *etcd) Update() error {
	return e.getRemoteConfig()
}

func (e *etcd) watch(ch chan<- struct{}) {
	rch := e.client.Watch(context.Background(), e.prefix, clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			e.logger.Info("config.etcd watcher triggered", zap.ByteString("key", ev.Kv.Key))
			select {
			case ch <- struct{}{}:
			default:
				e.logger.Info("config.etcd watcher response dropped")
			}
		}
	}
}
