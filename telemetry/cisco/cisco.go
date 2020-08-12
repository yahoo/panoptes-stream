package cisco

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/cisco/gnmi"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/cisco/mdt"
)

func Register(telemetryRegistrar *telemetry.Registrar) {
	telemetryRegistrar.Register("cisco.gnmi", gnmi.Version(), gnmi.New)
	telemetryRegistrar.Register("cisco.mdt", mdt.Version(), mdt.New)
}
