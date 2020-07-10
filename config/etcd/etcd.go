package etcd

import (
	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"go.etcd.io/etcd/clientv3"
	"go.uber.org/zap"
)

type etcd struct {
	client *clientv3.Client

	filename  string
	devices   []config.Device
	producers []config.Producer
	databases []config.Database
	global    *config.Global

	informer chan struct{}

	logger *zap.Logger
}

type etcdConfig struct {
	Endpoints []string
}

// New creates an etcd configuration
func New(filename string) (config.Config, error) {
	var (
		cfg  = &etcdConfig{}
		etcd = &etcd{
			informer: make(chan struct{}, 1),
		}
	)
	if err := yaml.Read(filename, cfg); err != nil {
		return nil, err
	}

	return etcd, nil
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

	return nil
}
