package database

import (
	"context"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

// Factory is a function that returns a new instance of database
type Factory func(context.Context, config.Database, *zap.Logger, telemetry.ExtDSChan) Database

// Database represents a database
type Database interface {
	Start()
}
