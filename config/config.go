package config

type Config interface {
	Devices() []Device
	Producers() []Producer
	Databases() []Database
	Global() Global
}

type DeviceConfig struct {
	Host string
	Port int

	TLSCertFile string `yaml:"tlsCertFile"`
	TLSKeyFile  string `yaml:"tlsKeyFile"`
	CAFile      string `yaml:"caFile"`

	username string
	password string
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
}
