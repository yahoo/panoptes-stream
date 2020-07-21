package config

import (
	"bytes"
	"net/url"

	"go.uber.org/zap"
)

// MockConfig represents mock configuration
type MockConfig struct {
	MDevices   []Device
	MProducers []Producer
	MDatabases []Database
	MGlobal    *Global

	MInformer chan struct{}

	LogOutput *MemSink
}

// MemSink repesents memory destination for logging
type MemSink struct {
	*bytes.Buffer
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
	cfg := zap.NewDevelopmentConfig()
	m.LogOutput = &MemSink{new(bytes.Buffer)}
	zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return m.LogOutput, nil
	})

	cfg.OutputPaths = []string{"memory://"}
	cfg.Encoding = "json"

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger
}

func (s *MemSink) Close() error { return nil }
func (s *MemSink) Sync() error  { return nil }
