package juniper

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/gnmi"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/jti"
)

func Register() []string {
	var reg []string
	reg = append(reg, gnmi.Register())
	reg = append(reg, jti.Register())

	return reg
}
