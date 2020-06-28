package register

import (
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper"
)

func RegisterVendor() {
	juniper.Register()
}
