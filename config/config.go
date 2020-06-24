package config

type Config interface {
	Devices() []Device
	Global() Global
}

type DeviceConfig struct {
	Host string
	Port int

	TLSCert string
	TLSKey  string
	TLSCa   string

	username string
	password string
}

type Sensor struct {
	Service  string
	Origin   string
	Path     string
	Mode     string
	Interval uint64
}

type Device struct {
	DeviceConfig

	Sensors []*Sensor
}

type Global struct {
	Redial int
}
