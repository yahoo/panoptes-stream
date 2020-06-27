package kafka

import (
	"log"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

type Kafka struct {
	ch telemetry.ExtDSChan
}

func Register() {
	log.Println("kafka register")
	producer.Register("kafka", New)
}

func New(cfg config.Producer, inChan telemetry.ExtDSChan) producer.Producer {
	return &Kafka{ch: inChan}
}

func (k *Kafka) Start() {
	for {
		v, ok := <-k.ch
		if !ok {
			break
		}

		v.DS.PrettyPrint()
	}
}
