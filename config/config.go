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

	TLSConfig TLSConfig
	Username  string
	Password  string
}

type Sensor struct {
	Service  string
	Output   string
	Disabled bool

	Origin            string
	Path              string
	Mode              string
	SampleInterval    uint64 `yaml:"sampleInterval"`
	HeartbeatInterval uint64 `yaml:"heartbeatInterval"`
	SuppressRedundant bool   `yaml:"suppressRedundant"`
}

type Device struct {
	DeviceConfig

	Sensors map[string][]*Sensor
}

type Producer struct {
	Name    string
	Service string
	Config  map[string]interface{}
}

type Database struct {
	Name    string
	Service string
	Config  map[string]interface{}
}

type Global struct {
	Version string
	Redial  int

	Discovery Discovery
	TLSConfig TLSConfig
	Status    Status
	Shard     Shard
	Logger    map[string]interface{}
}

type TLSConfig struct {
	Disabled bool

	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	CertFile           string `yaml:"certFile"`
	KeyFile            string `yaml:"keyFile"`
	CAFile             string `yaml:"caFile"`
}

type Status struct {
	Addr     string
	Disabled bool
}

type Shard struct {
	Enabled           bool
	InitializingShard int
	NumberOfNodes     int
}

type Discovery struct {
	Service    string `yaml:"service"`
	Prefix     string `yaml:"prefix"`
	ConfigFile string `yaml:"configFile"`
}
