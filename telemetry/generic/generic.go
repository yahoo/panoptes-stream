package generic

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/generic/gnmi"
)

func Register(telemetryRegistrar *telemetry.Registrar) {
	telemetryRegistrar.Register("generic.gnmi", gnmi.Version(), gnmi.New)
}
