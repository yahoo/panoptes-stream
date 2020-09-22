package mqueue

import (
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue/kafka"
)

// Register registers producers to producer registrar
func Register(producerRegistrar *producer.Registrar) {
	producerRegistrar.Register("kafka", "segment.io", kafka.New)
	producerRegistrar.Register("nsq", "nsq.io", kafka.New)
}
