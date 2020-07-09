package arista

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/arista/gnmi"
)

func Register(telemetryRegistrar *telemetry.Registrar) {
	telemetryRegistrar.Register("arista.gnmi", gnmi.Version(), gnmi.New)
}
