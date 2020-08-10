package cisco

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/cisco/gnmi"
)

func Register(telemetryRegistrar *telemetry.Registrar) {
	telemetryRegistrar.Register("cisco.gnmi", gnmi.Version(), gnmi.New)
}
