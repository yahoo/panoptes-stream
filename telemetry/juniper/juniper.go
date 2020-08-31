//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package juniper

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/gnmi"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/jti"
)

// Register Juniper telemetries
func Register(telemetryRegistrar *telemetry.Registrar) {
	telemetryRegistrar.Register("juniper.gnmi", gnmi.Version(), gnmi.New)
	telemetryRegistrar.Register("juniper.jti", jti.Version(), jti.New)
}
