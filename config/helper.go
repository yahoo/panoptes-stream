package config

import (
	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type DeviceTemplate struct {
	DeviceConfig `yaml:",inline"`

	Sensors []string
}

func ConvDeviceTemplate(d DeviceTemplate) Device {
	device := Device{}
	b, _ := json.Marshal(&d)
	json.Unmarshal(b, &device)
	device.Sensors = nil
	return device
}

func GetLogger(lcfg map[string]interface{}) *zap.Logger {
	var cfg zap.Config
	b, err := json.Marshal(lcfg)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil
	}

	cfg.Encoding = "console"
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = nil
	cfg.DisableStacktrace = true

	logger, err := cfg.Build()
	if err != nil {
		return nil
	}

	return logger
}

func GetDefaultLogger() *zap.Logger {
	var cfg = zap.Config{
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		Encoding:         "console",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = nil
	cfg.DisableStacktrace = true

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
