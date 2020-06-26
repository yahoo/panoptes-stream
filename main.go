package main

import (
	"context"
	"log"
	"strings"
	"sync"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/vendor"
	"google.golang.org/grpc"
	//log "github.com/golang/glog"
)

func main() {
	vendor.Register()
	mqueue.Register()

	outChan := make(telemetry.ExtDSChan, 1)
	y := yaml.LoadConfig("etc/config.yaml")

	dp := dispatcher{chMap: make(map[string]telemetry.ExtDSChan)}
	dp.prepare(y)
	go dp.dispatcher(outChan)

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

type dispatcher struct {
	chMap map[string]telemetry.ExtDSChan
}

func (d *dispatcher) prepare(cfg config.Config) {
	// TODO same proc for db
	// create channel map
	for _, p := range cfg.Producers() {
		mqNew, ok := producer.GetProducerFactory(p.Service)
		if !ok {
			// TODO
			continue
		}

		// register channel
		d.chMap[p.Name] = make(telemetry.ExtDSChan, 1)
		// construct
		m := mqNew(p, d.chMap[p.Name])
		// start the producer
		go m.Start()
	}
}

func (d *dispatcher) dispatcher(ch telemetry.ExtDSChan) {
	for {
		extDS, _ := <-ch
		output := strings.Split(extDS.Output, "::")
		if len(output) < 2 {
			continue
		}

		d.chMap[output[0]] <- extDS

	}
}
