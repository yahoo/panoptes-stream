//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package config

import (
	"go.uber.org/zap"
)

// Config represets panoptes configuration
type Config interface {
	Devices() []Device
	Producers() []Producer
	Databases() []Database
	Sensors() []Sensor
	Global() *Global
	Informer() chan struct{}
	Logger() *zap.Logger
	Update() error
}

// DeviceConfig represents device configuration
type DeviceConfig struct {
	Host string
	Port int

	GroupID int `yaml:"groupID"`

	DeviceOptions `yaml:",inline"`
}

// Sensor represents telemetry sensor
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

// Device represents device configuration with sensors
type Device struct {
	DeviceConfig

	Sensors map[string][]*Sensor
}

// Producer represents producer configuration
type Producer struct {
	Name    string
	Service string
	Config  interface{}
}

// Database represents database configuration
type Database struct {
	Name    string
	Service string
	Config  interface{}
}

// Global represents global configuration
type Global struct {
	Discovery        Discovery
	Status           Status
	Shards           Shards
	DeviceOptions    DeviceOptions `yaml:"deviceOptions"`
	WatcherDisabled  bool          `yaml:"watcherDisabled"`
	BufferSize       int           `yaml:"bufferSize"`
	OutputBufferSize int           `yaml:"outputBufferSize"`
	Version          string
	Logger           map[string]interface{}
	Dialout          Dialout
}

// TLSConfig represents TLS client configuration
type TLSConfig struct {
	Enabled bool

	CertFile           string `yaml:"certFile"`
	KeyFile            string `yaml:"keyFile"`
	CAFile             string `yaml:"caFile"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
}

// Status represents status: self monitoring and healthcheck configuration
type Status struct {
	Addr      string
	Disabled  bool
	TLSConfig TLSConfig `yaml:"tlsConfig"`
}

// Shards represents shard service configuration
type Shards struct {
	Enabled            bool
	InitializingShards int `yaml:"initializingShards"`
	MinimumShards      int `yaml:"minimumShards"`
	NumberOfNodes      int `yaml:"numberOfNodes"`
}

// Discovery represents discovery service
type Discovery struct {
	Service string
	Config  interface{}
}

// DeviceOptions represents global device options
type DeviceOptions struct {
	TLSConfig TLSConfig `yaml:"tlsConfig"`
	Username  string
	Password  string
	Timeout   int
}

// Dialout represents dialout service
type Dialout struct {
	TLSConfig     TLSConfig `yaml:"tlsConfig"`
	DefaultOutput string    `yaml:"defaultOutput"`
	Services      map[string]DialoutService
}

// DialoutService represent specific dialout telemetry
type DialoutService struct {
	Addr    string
	Workers int
}

// DeviceTemplate represents device configuration structure
type DeviceTemplate struct {
	DeviceConfig `yaml:",inline"`

	Sensors []string
}
