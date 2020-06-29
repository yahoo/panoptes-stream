package register

import (
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/producer/console"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper"
)

func Telemetry(telemetryRegistrar *telemetry.Registrar) {
	juniper.Register(telemetryRegistrar)
}

func Producer(producerRegistrar *producer.Registrar) {
	mqueue.Register(producerRegistrar)
	console.Register(producerRegistrar)
}
