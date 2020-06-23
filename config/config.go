package config

type Config interface {
	Get() []Devices
}

type Device struct {
	Host    string
	Sensors []Sensor
}

type Sensor struct {
	Service  string
	Origin   string
	Path     string
	Mode     string
	Interval uint64
}
