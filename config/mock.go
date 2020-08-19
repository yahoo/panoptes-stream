package config

import (
	"bytes"
	"encoding/json"
	"net/url"

	"go.uber.org/zap"
)

// MockConfig represents mock configuration
type MockConfig struct {
	MDevices   []Device
	MProducers []Producer
	MDatabases []Database
	MSensors   []Sensor
	MGlobal    *Global

	MInformer chan struct{}

	LogOutput *MemSink

	logger *zap.Logger
}

// MemSink repesents memory destination for logging
type MemSink struct {
	*bytes.Buffer
}

// NewMockConfig constructs mock configuration
// it writes logs to memory and accessable from LogOutput.
func NewMockConfig() *MockConfig {
	var (
		err error
		m   = &MockConfig{}
	)

	cfg := zap.NewDevelopmentConfig()
	m.LogOutput = &MemSink{new(bytes.Buffer)}
	zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return m.LogOutput, nil
	})

	cfg.OutputPaths = []string{"memory://"}
	cfg.DisableStacktrace = true
	cfg.Encoding = "json"

	m.logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}

	return m
}

// Devices returns configured devices
func (m *MockConfig) Devices() []Device {
	return m.MDevices
}

// Producers returns configured producers
func (m *MockConfig) Producers() []Producer {
	return m.MProducers
}

// Databases returns configured databases
func (m *MockConfig) Databases() []Database {
	return m.MDatabases
}

// Sensors returns configured sensors
func (m *MockConfig) Sensors() []Sensor {
	return m.MSensors
}

// Global returns global configuration
func (m *MockConfig) Global() *Global {
	return m.MGlobal
}

// Informer returns informer channel
func (m *MockConfig) Informer() chan struct{} {
	return m.MInformer
}

// Update doesn't do anything
func (m *MockConfig) Update() error {
	return nil
}
func (m *MockConfig) Logger() *zap.Logger {
	return m.logger
}

func (s *MemSink) Close() error { return nil }
func (s *MemSink) Sync() error  { return nil }
func (s *MemSink) Unmarshal() map[string]string {
	defer s.Reset()
	v := make(map[string]string)
	json.Unmarshal(s.Bytes(), &v)
	return v
}
