package kafka

import (
	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

type Kafka struct {
	ch telemetry.ExtDSChan
}

func New(cfg config.Producer, lg *zap.Logger, inChan telemetry.ExtDSChan) producer.Producer {
	return &Kafka{ch: inChan}
}

func (k *Kafka) Start() {
	for {
		v, ok := <-k.ch
		if !ok {
			break
		}

		v.DS.PrettyPrint("stdout")
	}
}
