//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package database

import (
	"context"

	"go.uber.org/zap"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/telemetry"
)

// Factory is a function that returns a new instance of database
type Factory func(context.Context, config.Database, *zap.Logger, telemetry.ExtDSChan) Database

// Database represents a database
type Database interface {
	Start()
}
