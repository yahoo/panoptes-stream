//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package juniper

import (
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/juniper/gnmi"
	"github.com/yahoo/panoptes-stream/telemetry/juniper/jti"
)

// Register Juniper telemetries
func Register(telemetryRegistrar *telemetry.Registrar) {
	telemetryRegistrar.Register("juniper.gnmi", gnmi.Version(), gnmi.New)
	telemetryRegistrar.Register("juniper.jti", jti.Version(), jti.New)
}
