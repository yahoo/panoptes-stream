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

	Username string
	Password string
}

type Sensor struct {
	Service string
	Output  string

	Origin   string
	Path     string
	Mode     string
	Interval uint64
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
	Redial int

	TLSConfig TLSConfig
	Status    Status
	Logger    map[string]interface{}
}

type TLSConfig struct {
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	CertFile           string `yaml:"certFile"`
	KeyFile            string `yaml:"keyFile"`
	CAFile             string `yaml:"caFile"`
}

type Status struct {
	Addr     string
	Disabled bool
}
