package yaml

import (
	"io/ioutil"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	yml "gopkg.in/yaml.v3"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

// yaml represents yaml configuration management
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

type producer struct {
	Service    string `yaml:"service"`
	ConfigFile string `yaml:"configFile"`
}

type database struct {
	Service    string `yaml:"service"`
	ConfigFile string `yaml:"configFile"`
}

type yamlConfig struct {
	Devices   []config.DeviceTemplate
	Sensors   map[string]config.Sensor
	Producers map[string]producer
	Databases map[string]database

	config.Global `yaml:",inline"`
}

// New constructs yaml configuration management
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

	y.devices = y.getDevices(yamlCfg)
	y.producers = y.getProducers(yamlCfg.Producers)
	y.databases = y.getDatabases(yamlCfg.Databases)
	y.sensors = y.getSensors(yamlCfg.Sensors)
	y.global = y.getGlobal(&yamlCfg.Global)

	if !yamlCfg.Global.WatcherDisabled {
		go func() {
			if err := y.watcher(); err != nil {
				y.logger.Error("watcher", zap.Error(err))
				os.Exit(1)
			}

		}()
	}

	return y, nil
}

// Update reads yaml file
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

// Devices returns configured devices
func (y *yaml) Devices() []config.Device {
	return y.devices
}

// Global returns global configuration
func (y *yaml) Global() *config.Global {
	return y.global
}

// Producers returns configured producers
func (y *yaml) Producers() []config.Producer {
	return y.producers
}

// Databases returns configured databases
func (y *yaml) Databases() []config.Database {
	return y.databases
}

// Sensors returns configured sensors
func (y *yaml) Sensors() []config.Sensor {
	return y.sensors
}

// Logger returns logging handler
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

// Read reads a file and deserialization data
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

func (y *yaml) getProducers(p map[string]producer) []config.Producer {
	var (
		producers []config.Producer
		cfg       interface{}
	)

	for name, pConfig := range p {
		if err := Read(pConfig.ConfigFile, &cfg); err != nil {
			y.logger.Error("yaml", zap.Error(err), zap.String("file", pConfig.ConfigFile))
			os.Exit(1)
		}

		producers = append(producers, config.Producer{
			Name:    name,
			Service: pConfig.Service,
			Config:  cfg,
		})
	}

	// built-in producer
	producers = append(producers, config.Producer{
		Name:    "console",
		Service: "console",
	})

	return producers
}

func (y *yaml) getDatabases(p map[string]database) []config.Database {
	var (
		databases []config.Database
		cfg       interface{}
	)

	for name, pConfig := range p {
		if err := Read(pConfig.ConfigFile, &cfg); err != nil {
			y.logger.Error("yaml", zap.Error(err), zap.String("file", pConfig.ConfigFile))
			os.Exit(1)
		}

		databases = append(databases, config.Database{
			Name:    name,
			Service: pConfig.Service,
			Config:  cfg,
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
	var conf = make(map[string]interface{})

	if g.Discovery.ConfigFile != "" {
		if err := Read(g.Discovery.ConfigFile, &conf); err != nil {
			y.logger.Error("yaml", zap.Error(err), zap.String("file", g.Discovery.ConfigFile))
			os.Exit(1)
		}

		g.Discovery.Config = conf
	}

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
