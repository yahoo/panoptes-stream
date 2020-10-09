package mqueue

import (
	"github.com/yahoo/panoptes-stream/producer"
	"github.com/yahoo/panoptes-stream/producer/mqueue/kafka"
)

// Register registers producers to producer registrar
func Register(producerRegistrar *producer.Registrar) {
	producerRegistrar.Register("kafka", "segment.io", kafka.New)
	producerRegistrar.Register("nsq", "nsq.io", kafka.New)
}
