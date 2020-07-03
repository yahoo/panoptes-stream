package producer

import (
	"context"

	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

type ProducerFactory func(context.Context, config.Producer, *zap.Logger, telemetry.ExtDSChan) Producer

type Producer interface {
	Start()
}
