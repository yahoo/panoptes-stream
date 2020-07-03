package database

import (
	"context"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

type DatabaseFactory func(context.Context, config.Database, *zap.Logger, telemetry.ExtDSChan) Database

type Database interface {
	Start()
}
