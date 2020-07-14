package yaml

import (
	"io/ioutil"
	"log"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	yml "gopkg.in/yaml.v2"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

type yaml struct {
	filename  string
	devices   []config.Device
	producers []config.Producer
	databases []config.Database
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

// New constructs new yaml config
func New(filename string) (config.Config, error) {
	yamlCfg := &yamlConfig{}
	if err := Read(filename, yamlCfg); err != nil {
		return &yaml{}, err
	}

	y := &yaml{
		filename: filename,

		devices:   configDevices(yamlCfg),
		producers: configProducers(yamlCfg.Producers),
		databases: configDatabases(yamlCfg.Databases),
		global:    &yamlCfg.Global,

		logger: config.GetLogger(yamlCfg.Global.Logger),

		informer: make(chan struct{}, 1),
	}

	go y.watcher()

	return y, nil
}

func (y *yaml) Update() error {
	yamlCfg := &yamlConfig{}

	if err := Read(y.filename, yamlCfg); err != nil {
		return err
	}

	y.devices = configDevices(yamlCfg)
	y.producers = configProducers(yamlCfg.Producers)
	y.databases = configDatabases(yamlCfg.Databases)
	y.global = &yamlCfg.Global

	return nil
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

func (y *yaml) Databases() []config.Database {
	return y.databases
}

func (y *yaml) Logger() *zap.Logger {
	return y.logger
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

	// add built-in producer
	producers = append(producers, config.Producer{
		Name:    "console",
		Service: "console",
	})

	return producers
}

func configDatabases(p map[string]database) []config.Database {
	var databases []config.Database
	for name, pConfig := range p {
		cfg := make(map[string]interface{})

		if name != "console" {
			if err := Read(pConfig.ConfigFile, &cfg); err != nil {
				log.Fatal(err)
			}
		}

		databases = append(databases, config.Database{
			Name:    name,
			Service: pConfig.Service,
			Config:  cfg,
		})
	}

	return databases
}

func (y *yaml) watcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
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

					y.logger.Info("watcher", zap.String("name", event.Name))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(y.filename)
	if err != nil {
		log.Fatal(err)
	}

	<-done
}

func (y *yaml) Informer() chan struct{} {
	return y.informer
}
