package vendor

import (
	jgnmi "juniper/gnmi"
	"juniper/jti"
)

func Register() {
	jgnmi.Register()
	jti.Register()
}
