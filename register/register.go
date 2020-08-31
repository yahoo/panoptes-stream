//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package register

import (
	"git.vzbuilders.com/marshadrad/panoptes/database"
	"git.vzbuilders.com/marshadrad/panoptes/database/tsdb"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/producer/console"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/arista"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/cisco"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper"
)

// Telemetry registers all available telemetries
func Telemetry(telemetryRegistrar *telemetry.Registrar) {
	juniper.Register(telemetryRegistrar)
	cisco.Register(telemetryRegistrar)
	arista.Register(telemetryRegistrar)
}

// Producer registers all available producers
func Producer(producerRegistrar *producer.Registrar) {
	mqueue.Register(producerRegistrar)
	console.Register(producerRegistrar)
}

// Database registers all available databases
func Database(databaseRegistrar *database.Registrar) {
	tsdb.Register(databaseRegistrar)
}
