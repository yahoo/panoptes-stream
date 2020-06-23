package telemetry

import (
	"fmt"
)

var R = make(map[string]string)

type Telemetry interface {
	Subscribe()
}

func Register(t Telemetry) {
	fmt.Println("hello")
	R["Q"] = "B"
}
