package register

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper"
)

func RegisterVendor(telemetryRegistrar *telemetry.TelemetryRegistrar) {
	juniper.Register(telemetryRegistrar)
}
