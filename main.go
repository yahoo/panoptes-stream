package main

import (
	"context"
	"log"
	"sync"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/vendor"
	"google.golang.org/grpc"
	//log "github.com/golang/glog"
)

func main() {
	vendor.Register()
	mqueue.Register()

	y := yaml.LoadConfig("etc/config.yaml")
	log.Printf("Config %#v", y)

	wg := sync.WaitGroup{}
	for _, device := range y.Devices() {
		log.Println(device.Host)
		conn, err := grpc.Dial(device.Host, grpc.WithInsecure(), grpc.WithUserAgent("Panoptes"))
		if err != nil {
			// TODO
			log.Fatal(err)
		}

		wg.Add(1)
		go func(device config.Device) {
			defer wg.Done()
			for sName, sensors := range device.Sensors {
				log.Println(sName, sensors)
				outChan := make(telemetry.DSChan, 1)
				nmiNew := telemetry.GetNMIFactory(sName)
				nmi := nmiNew(conn, sensors, outChan)
				err := nmi.Start(context.Background())
				if err != nil {
					log.Println(err)
				}
			}
		}(device)

	}

	wg.Wait()

}
