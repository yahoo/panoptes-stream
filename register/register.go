package register

import (
	"git.vzbuilders.com/marshadrad/panoptes/database"
	"git.vzbuilders.com/marshadrad/panoptes/database/tsdb"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/producer/console"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/arista"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper"
)

func Telemetry(telemetryRegistrar *telemetry.Registrar) {
	juniper.Register(telemetryRegistrar)
	arista.Register(telemetryRegistrar)
}

func Producer(producerRegistrar *producer.Registrar) {
	mqueue.Register(producerRegistrar)
	console.Register(producerRegistrar)
}

func Database(databaseRegistrar *database.Registrar) {
	tsdb.Register(databaseRegistrar)
}
