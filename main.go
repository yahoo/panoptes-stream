package main

import (
	"log"
	"sync"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/vendor"
	"google.golang.org/grpc"
	//log "github.com/golang/glog"
)

func main() {
	vendor.Register()

	log.Println("panoptes", telemetry.R)

	y := yaml.LoadConfig()
	wg := sync.WaitGroup{}
	for _, device := range y.Devices() {
		log.Println(device.Host)
		conn, err := grpc.Dial(device.Host)
		if err != nil {
			// TODO
			log.Fatal(err)
		}

		wg.Add(1)
		go func(device config.Device) {
			defer wg.Done()
			for sName, sensors := range device.Sensors {
				log.Println(sName, sensors)
				outChan := make(telemetry.KVChan, 1)
				f := telemetry.R[sName]
				f(conn, sensors, outChan)
			}
		}(device)

	}

	wg.Wait()

}
