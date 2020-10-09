//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package etcd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/config/yaml"
	"github.com/yahoo/panoptes-stream/secret"
)

// etcd represents etcd configuration management.
type etcd struct {
	client *clientv3.Client

	prefix    string
	devices   []config.Device
	producers []config.Producer
	databases []config.Database
	sensors   []config.Sensor
	global    *config.Global

	informer chan struct{}

	logger *zap.Logger
}

type etcdConfig struct {
	Endpoints   []string
	Prefix      string
	DialTimeout int

	TLSConfig config.TLSConfig
}

// New constructs etcd configuration management.
func New(filename string) (config.Config, error) {
	var (
		err       error
		tlsConfig *tls.Config
	)

	etcd := &etcd{
		informer: make(chan struct{}, 1),
		global:   &config.Global{},
		logger:   config.GetDefaultLogger(),
	}

	config := &etcdConfig{}

	err = yaml.Read(filename, config)
	if err != nil && filename != "-" {
		etcd.logger.Fatal("etcd", zap.Error(err))
	}

	prefix := "panoptes_config_etcd"
	err = envconfig.Process(prefix, config)
	if err != nil {
		return nil, err
	}

	setDefaultConfig(config)

	etcd.prefix = config.Prefix

	if config.TLSConfig.Enabled {
		tlsConfig, err = secret.GetTLSConfig(&config.TLSConfig)
		if err != nil {
			return nil, err
		}
	}

	etcd.client, err = clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: time.Duration(config.DialTimeout) * time.Second,
		TLS:         tlsConfig,
		LogConfig: &zap.Config{
			Development:      false,
			Level:            zap.NewAtomicLevelAt(zap.ErrorLevel),
			Encoding:         "json",
			EncoderConfig:    zap.NewProductionEncoderConfig(),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		},
	})
	if err != nil {
		return nil, err
	}

	if err = etcd.getRemoteConfig(); err != nil {
		return nil, err
	}

	etcd.logger.Info("panoptes-stream", zap.String("version", etcd.global.Version))
	etcd.logger.Info("panoptes-stream", zap.String("go version", runtime.Version()), zap.String("go os/arch", runtime.GOOS+"/"+runtime.GOARCH))

	if !etcd.global.WatcherDisabled {
		go etcd.watch(etcd.informer)
	} else {
		go etcd.signalHandler()
	}

	return etcd, nil
}

func (e *etcd) getRemoteConfig() error {
	var (
		devicesTpl = make(map[string]config.DeviceTemplate)
		sensors    = make(map[string]*config.Sensor)
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	resp, err := e.client.Get(ctx, e.prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return err
	}

	e.devices = e.devices[:0]
	e.producers = e.producers[:0]
	e.databases = e.databases[:0]
	e.sensors = e.sensors[:0]

	if len(resp.Kvs) < 1 {
		return fmt.Errorf("etcd config is not exist , prefix=%s", e.prefix)
	}

	for _, ev := range resp.Kvs {
		key := strings.TrimPrefix(string(ev.Key), e.prefix)
		folder, k := path.Split(key)

		if len(key) < 1 {
			return fmt.Errorf("etcd is empty, prefix=%s", e.prefix)
		}

		if !json.Valid(ev.Value) {
			return fmt.Errorf("invalid JSON encoding - %s", key)
		}

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
			if err := config.SensorValidation(sensor); err != nil {
				e.logger.Error("etcd", zap.Error(err))
				continue
			}
			sensors[k] = &sensor
			e.sensors = append(e.sensors, sensor)
		default:
			if k == "global" {
				err = json.Unmarshal(ev.Value, &e.global)
				if err != nil {
					return err
				}

				prefix := "panoptes"
				err = envconfig.Process(prefix, e.global)
				if err != nil {
					return err
				}

				config.SetDefaultGlobal(e.global)

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

		if err := config.DeviceValidation(device); err != nil {
			e.logger.Error("etcd", zap.Error(err))
			continue
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

// Devices returns configured devices.
func (e *etcd) Devices() []config.Device {
	return e.devices
}

// Producers returns configured producers.
func (e *etcd) Producers() []config.Producer {
	return e.producers
}

// Databases returns configured databases.
func (e *etcd) Databases() []config.Database {
	return e.databases
}

// Sensors returns configured sensors.
func (e *etcd) Sensors() []config.Sensor {
	return e.sensors
}

// Global returns global configuration.
func (e *etcd) Global() *config.Global {
	return e.global
}

// Informer returns informer channel.
func (e *etcd) Informer() chan struct{} {
	return e.informer
}

// Logger returns logging handler.
func (e *etcd) Logger() *zap.Logger {
	return e.logger
}

// Update gets configuration from etcd key value store.
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

func (e *etcd) signalHandler() {
	var signalCh = make(chan os.Signal, 1)

	signal.Notify(signalCh, syscall.SIGHUP)

	for {
		<-signalCh
		e.logger.Info("etcd.sighup", zap.String("event", "triggered"))

		select {
		case e.informer <- struct{}{}:
		default:
			e.logger.Warn("etcd.sighub", zap.String("event", "notification has been dropped"))
		}
	}
}

func setDefaultConfig(config *etcdConfig) {
	if len(config.Endpoints) < 1 {
		config.Endpoints = []string{"127.0.0.1:2379"}
	}

	if len(config.Prefix) < 1 {
		config.Prefix = "panoptes/config/"
	}

	if config.DialTimeout < 1 {
		config.DialTimeout = 2
	}
}
