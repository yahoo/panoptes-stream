//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package tsdb

import (
	"github.com/yahoo/panoptes-stream/database"
	"github.com/yahoo/panoptes-stream/database/tsdb/influxdb"
)

// Register registers databases to database registrar
func Register(databaseRegistrar *database.Registrar) {
	databaseRegistrar.Register("influxdb", "influxdata.com", influxdb.New)
}
