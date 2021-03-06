//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package cisco

import (
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/cisco/gnmi"
	"github.com/yahoo/panoptes-stream/telemetry/cisco/mdt"
)

// Register Cisco telemetries
func Register(telemetryRegistrar *telemetry.Registrar) {
	telemetryRegistrar.Register("cisco.gnmi", gnmi.Version(), gnmi.New)
	telemetryRegistrar.Register("cisco.mdt", mdt.Version(), mdt.New)
}
