package main

import (
	"encoding/json"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetLogger(lcfg map[string]interface{}) *zap.Logger {
	var cfg zap.Config
	b, err := json.Marshal(lcfg)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		panic(err)
	}

	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeCaller = nil

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
