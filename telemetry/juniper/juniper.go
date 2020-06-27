package juniper

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/gnmi"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/jti"
)

func Register() {
	gnmi.Register()
	jti.Register()
}
