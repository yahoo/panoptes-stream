//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package register

import (
	"github.com/yahoo/panoptes-stream/database"
	"github.com/yahoo/panoptes-stream/database/tsdb"
	"github.com/yahoo/panoptes-stream/producer"
	"github.com/yahoo/panoptes-stream/producer/console"
	"github.com/yahoo/panoptes-stream/producer/mqueue"
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/arista"
	"github.com/yahoo/panoptes-stream/telemetry/cisco"
	"github.com/yahoo/panoptes-stream/telemetry/juniper"
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
