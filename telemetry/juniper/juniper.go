package juniper

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/gnmi"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/jti"
)

func Register(telemetryRegistrar *telemetry.Registrar) {
	telemetryRegistrar.Register("juniper.gnmi", gnmi.Version(), gnmi.New)
	telemetryRegistrar.Register("juniper.jti", jti.Version(), jti.New)
}
