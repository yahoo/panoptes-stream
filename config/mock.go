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

func (m *MockConfig) Devices() []Device {
	return m.MDevices
}
func (m *MockConfig) Producers() []Producer {
	return m.MProducers
}
func (m *MockConfig) Databases() []Database {
	return m.MDatabases
}
func (m *MockConfig) Sensors() []Sensor {
	return m.MSensors
}
func (m *MockConfig) Global() *Global {
	return m.MGlobal
}
func (m *MockConfig) Informer() chan struct{} {
	return m.MInformer
}
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
