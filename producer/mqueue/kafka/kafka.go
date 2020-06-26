package kafka

import (
	"log"

	"git.vzbuilders.com/marshadrad/panoptes/producer"
)

type Kafka struct{}

func Register() {
	log.Println("kafka register")
}

func New() producer.Producer {
	return &Kafka{}
}

func (k *Kafka) Setup() {

}

func (k *Kafka) Start() {

}
