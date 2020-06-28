package mqueue

import (
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue/kafka"
)

func Register(producerRegistrar *producer.ProducerRegistrar) {
	producerRegistrar.Register("kafka", kafka.New)
}
