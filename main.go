package main

import (
	"context"
	"log"
	"strings"
	"sync"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"git.vzbuilders.com/marshadrad/panoptes/demux"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/vendor"
	"google.golang.org/grpc"
	//log "github.com/golang/glog"
)

func main() {
	ctx := context.Background()

	vendor.Register()
	mqueue.Register()

	outChan := make(telemetry.ExtDSChan, 1)
	cfg := yaml.LoadConfig("etc/config.yaml")

	dp := demux.New(cfg, outChan) //demux{chMap: make(map[string]telemetry.ExtDSChan)}
	//dp.prepare(y)
	//go dp.dispatcher(outChan)
	dp.Init()
	go dp.Start(ctx)

	wg := sync.WaitGroup{}
	for _, device := range cfg.Devices() {
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
				nmiNew := telemetry.GetNMIFactory(sName)
				nmi := nmiNew(conn, sensors, outChan)
				err := nmi.Start(context.Background())
				if err != nil {
					log.Println(err)
				}

				for _, sensor := range sensors {
					log.Printf(" SS %#v", strings.Split(sensor.Output, "::")[0])
				}
			}
		}(device)

	}

	// dispatch fun (outChan) to producer or database

	wg.Wait()

}
