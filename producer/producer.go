package producer

import (
	"context"

	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

// Factory is a function that returns a new instance of producer
type Factory func(context.Context, config.Producer, *zap.Logger, telemetry.ExtDSChan) Producer

// Producer represents a producer
type Producer interface {
	Start()
}
