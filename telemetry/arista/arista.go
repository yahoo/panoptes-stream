//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package arista

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/arista/gnmi"
)

// Register Arista telemetries
func Register(telemetryRegistrar *telemetry.Registrar) {
	telemetryRegistrar.Register("arista.gnmi", gnmi.Version(), gnmi.New)
}
