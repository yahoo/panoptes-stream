package config

type Config interface {
	Devices() []Device
	Producers() []Producer
	Global() *Global
	Informer() chan struct{}
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

type Global struct {
	TLSConfig

	Redial int
	Logger map[string]interface{}
}

type TLSConfig struct {
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
	TLSCertFile        string `yaml:"tlsCertFile"`
	TLSKeyFile         string `yaml:"tlsKeyFile"`
	CAFile             string `yaml:"caFile"`
}
