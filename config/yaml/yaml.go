//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package yaml

import (
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	yml "gopkg.in/yaml.v3"

	"github.com/yahoo/panoptes-stream/config"
)

// yaml represents yaml configuration management.
type yaml struct {
	filename  string
	devices   []config.Device
	producers []config.Producer
	databases []config.Database
	sensors   []config.Sensor
	global    *config.Global

	informer chan struct{}

	logger *zap.Logger
}

type yamlConfig struct {
	Devices   []config.DeviceTemplate
	Sensors   map[string]config.Sensor
	Producers map[string]config.Producer
	Databases map[string]config.Database

	config.Global `yaml:",inline"`
}

// New constructs yaml configuration management.
func New(filename string) (config.Config, error) {
	yamlCfg := &yamlConfig{}
	if err := Read(filename, yamlCfg); err != nil {
		return &yaml{}, err
	}

	y := &yaml{
		filename: filename,
		logger:   config.GetLogger(yamlCfg.Global.Logger),
		informer: make(chan struct{}, 1),
	}

	y.global = y.getGlobal(&yamlCfg.Global)

	y.logger.Info("panoptes-stream", zap.String("version", y.Global().Version))
	y.logger.Info("panoptes-stream", zap.String("go version", runtime.Version()), zap.String("go os/arch", runtime.GOOS+"/"+runtime.GOARCH))

	y.devices = y.getDevices(yamlCfg)
	y.producers = y.getProducers(yamlCfg.Producers)
	y.databases = y.getDatabases(yamlCfg.Databases)
	y.sensors = y.getSensors(yamlCfg.Sensors)

	if !yamlCfg.Global.WatcherDisabled {
		go func() {
			if err := y.watcher(); err != nil {
				y.logger.Fatal("watcher", zap.Error(err))
			}

		}()
	} else {
		go y.signalHandler()
	}

	return y, nil
}

// Update reads yaml file again and updates config.
func (y *yaml) Update() error {
	yamlCfg := &yamlConfig{}

	if err := Read(y.filename, yamlCfg); err != nil {
		return err
	}

	y.devices = y.getDevices(yamlCfg)
	y.producers = y.getProducers(yamlCfg.Producers)
	y.databases = y.getDatabases(yamlCfg.Databases)
	y.sensors = y.getSensors(yamlCfg.Sensors)
	y.global = y.getGlobal(&yamlCfg.Global)

	return nil
}

// Devices returns configured devices.
func (y *yaml) Devices() []config.Device {
	return y.devices
}

// Global returns global configuration.
func (y *yaml) Global() *config.Global {
	return y.global
}

// Producers returns configured producers.
func (y *yaml) Producers() []config.Producer {
	return y.producers
}

// Databases returns configured databases.
func (y *yaml) Databases() []config.Database {
	return y.databases
}

// Sensors returns configured sensors.
func (y *yaml) Sensors() []config.Sensor {
	return y.sensors
}

// Logger returns logging handler.
func (y *yaml) Logger() *zap.Logger {
	return y.logger
}

func (y *yaml) getDevices(cfg *yamlConfig) []config.Device {
	sensors := make(map[string]*config.Sensor)
	for name, sensor := range cfg.Sensors {
		if err := config.SensorValidation(sensor); err != nil {
			y.logger.Error("yaml", zap.Error(err))
			continue
		}

		sensor := sensor
		sensors[name] = &sensor
	}

	devices := []config.Device{}
	for _, d := range cfg.Devices {

		device := config.ConvDeviceTemplate(d)
		device.Sensors = make(map[string][]*config.Sensor)

		for _, s := range d.Sensors {
			sensor, ok := sensors[s]
			if !ok {
				y.logger.Error("yaml", zap.String("msg", "sensor not exist"), zap.String("sensor", s))
				continue
			}

			if !sensor.Disabled {
				device.Sensors[sensor.Service] = append(device.Sensors[sensor.Service], sensor)
			}
		}

		if err := config.DeviceValidation(device); err != nil {
			y.logger.Error("yaml", zap.Error(err))
			continue
		}

		devices = append(devices, device)
	}

	return devices
}

// Read reads a file and deserialization data.
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

func (y *yaml) getProducers(p map[string]config.Producer) []config.Producer {
	var producers []config.Producer

	for name, pConfig := range p {
		producers = append(producers, config.Producer{
			Name:    name,
			Service: pConfig.Service,
			Config:  pConfig.Config,
		})
	}

	// built-in producer
	producers = append(producers, config.Producer{
		Name:    "console",
		Service: "console",
	})

	return producers
}

func (y *yaml) getDatabases(d map[string]config.Database) []config.Database {
	var databases []config.Database

	for name, dConfig := range d {
		databases = append(databases, config.Database{
			Name:    name,
			Service: dConfig.Service,
			Config:  dConfig.Config,
		})
	}

	return databases
}

func (y *yaml) getSensors(s map[string]config.Sensor) []config.Sensor {
	var sensors []config.Sensor
	for _, sensor := range s {
		if err := config.SensorValidation(sensor); err != nil {
			continue
		}

		sensors = append(sensors, sensor)
	}

	return sensors
}

func (y *yaml) getGlobal(g *config.Global) *config.Global {
	envconfig.Process("panoptes", g)

	config.SetDefaultGlobal(g)

	return g
}

func (y *yaml) watcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					select {
					case y.informer <- struct{}{}:
					default:
					}

					y.logger.Info("watcher.loop", zap.String("name", event.Name))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}

				y.logger.Error("watcher.loop", zap.Error(err))
			}
		}
	}()

	err = watcher.Add(y.filename)
	if err != nil {
		return err
	}

	<-done

	return nil
}

func (y *yaml) Informer() chan struct{} {
	return y.informer
}

func (y *yaml) signalHandler() {
	var signalCh = make(chan os.Signal, 1)

	signal.Notify(signalCh, syscall.SIGHUP)

	for {
		<-signalCh
		y.logger.Info("yaml.sighup", zap.String("event", "triggered"))

		select {
		case y.informer <- struct{}{}:
		default:
			y.logger.Warn("yaml.sighub", zap.String("event", "notification has been dropped"))
		}
	}
}
