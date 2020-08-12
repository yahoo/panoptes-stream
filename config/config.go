package config

import "go.uber.org/zap"

type Config interface {
	Devices() []Device
	Producers() []Producer
	Databases() []Database
	Global() *Global
	Informer() chan struct{}
	Logger() *zap.Logger
	Update() error
}

type DeviceConfig struct {
	Host string
	Port int

	GroupID int

	DeviceOptions `yaml:",inline"`
}

type Sensor struct {
	Service  string
	Output   string
	Disabled bool

	Origin            string
	Path              string
	Mode              string
	SampleInterval    int  `yaml:"sampleInterval"`
	HeartbeatInterval int  `yaml:"heartbeatInterval"`
	SuppressRedundant bool `yaml:"suppressRedundant"`

	Subscription string
}

type Device struct {
	DeviceConfig

	Sensors map[string][]*Sensor
}

type Producer struct {
	Name    string
	Service string
	Config  interface{}
}

type Database struct {
	Name    string
	Service string
	Config  interface{}
}

type Global struct {
	Discovery       Discovery
	Status          Status
	Shard           Shard
	DeviceOptions   DeviceOptions `yaml:"deviceOptions"`
	WatcherDisabled bool          `yaml:"watcherDisabled"`
	Version         string
	Logger          map[string]interface{}
}

type TLSConfig struct {
	Enabled bool

	CertFile           string `yaml:"certFile"`
	KeyFile            string `yaml:"keyFile"`
	CAFile             string `yaml:"caFile"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
}

type Status struct {
	Addr      string
	Disabled  bool
	TLSConfig TLSConfig `yaml:"tlsConfig"`
}

type Shard struct {
	Enabled           bool
	InitializingShard int `yaml:"initializingShard"`
	NumberOfNodes     int `yaml:"numberOfNodes"`
}

type Discovery struct {
	ConfigFile string `yaml:"configFile"`
	Service    string
	Config     interface{}
}

type DeviceOptions struct {
	TLSConfig TLSConfig `yaml:"tlsConfig"`
	Username  string
	Password  string
}
