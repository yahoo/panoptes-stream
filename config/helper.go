package config

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ConvDeviceTemplate transforms devicetemplate to device
func ConvDeviceTemplate(d DeviceTemplate) Device {
	device := Device{}
	b, _ := json.Marshal(&d)
	json.Unmarshal(b, &device)
	device.Sensors = nil
	return device
}

// GetLogger tries to create a zap logger based on the user configuration
func GetLogger(lcfg map[string]interface{}) *zap.Logger {
	var cfg zap.Config
	b, err := json.Marshal(lcfg)
	if err != nil {
		return GetDefaultLogger()
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return GetDefaultLogger()
	}

	cfg.Encoding = "console"
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = nil
	cfg.DisableStacktrace = true

	logger, err := cfg.Build()
	if err != nil {
		return GetDefaultLogger()
	}

	return logger
}

// GetDefaultLogger creates default zap logger
func GetDefaultLogger() *zap.Logger {
	var cfg = zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = nil
	cfg.DisableStacktrace = true

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}

// DeviceValidation validates configured device
func DeviceValidation(device Device) error {
	if len(device.Sensors) < 1 {
		return fmt.Errorf("device: %s doesn't have any sensors", device.Host)
	}

	if len(device.Host) < 1 {
		return fmt.Errorf("device: %s doesn't have host", device.Host)
	}

	if device.Port < 1 {
		return fmt.Errorf("device: %s has invalid port", device.Host)
	}

	return nil
}

// SensorValidation validates configured sensor
func SensorValidation(sensor Sensor) error {
	availSensors := map[string]bool{
		"arista.gnmi":       true,
		"juniper.gnmi":      true,
		"cisco.gnmi":        true,
		"cisco.mdt":         true,
		"cisco.mdt.dialout": true,
		"juniper.jti":       true,
	}

	if _, ok := availSensors[sensor.Service]; !ok {
		return fmt.Errorf("sensor:%s not available", sensor.Service)
	}

	return nil
}

// SetDefaultGlobal set global default value
func SetDefaultGlobal(g *Global) {
	g.Version = "0.0.1"

	if g.BufferSize == 0 {
		g.BufferSize = 20000
	}

	if g.OutputBufferSize == 0 {
		g.OutputBufferSize = 10000
	}
}
