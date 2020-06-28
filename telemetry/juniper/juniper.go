package juniper

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/gnmi"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/jti"
)

func Register(telemetryRegistrar *telemetry.TelemetryRegistrar) {
	telemetryRegistrar.Register("juniper.gnmi", gnmi.New)
	telemetryRegistrar.Register("juniper.jti", jti.New)
}
