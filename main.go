package main

import (
	"log"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/vendor"
)

func main() {
	vendor.Register()

	log.Println("hello world", telemetry.R)
}
