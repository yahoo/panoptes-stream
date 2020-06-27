package main

import (
	"context"
	"log"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/config/yaml"
	"git.vzbuilders.com/marshadrad/panoptes/demux"
	"git.vzbuilders.com/marshadrad/panoptes/producer/mqueue"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/register"
	"google.golang.org/grpc"
	//log "github.com/golang/glog"
)

func main() {
	ctx := context.Background()

	register.Vendor()
	mqueue.Register()

	outChan := make(telemetry.ExtDSChan, 1)
	cfg := yaml.LoadConfig("etc/config.yaml")

	dp := demux.New(cfg, outChan)
	dp.Init()
	go dp.Start(ctx)

	p := NewPanoptes(ctx, outChan)
	for _, device := range cfg.Devices() {
		p.subscribe(device)
	}

	<-ctx.Done()
}

type panoptes struct {
	register map[string]context.CancelFunc
	ctx      context.Context
	outChan  telemetry.ExtDSChan
}

func NewPanoptes(ctx context.Context, outChan telemetry.ExtDSChan) *panoptes {
	return &panoptes{
		register: make(map[string]context.CancelFunc),
		ctx:      ctx,
		outChan:  outChan,
	}
}

func (p *panoptes) subscribe(device config.Device) {
	var ctx context.Context
	ctx, p.register[device.Host] = context.WithCancel(p.ctx)

	for sName, sensors := range device.Sensors {
		go func(sName string, sensors []*config.Sensor) {
			for {
				conn, err := grpc.Dial(device.Host, grpc.WithInsecure(), grpc.WithUserAgent("Panoptes"))
				if err != nil {
					log.Fatal(err)
				}

				NewNMI := telemetry.GetNMIFactory(sName)
				nmi := NewNMI(conn, sensors, p.outChan)
				err = nmi.Start(ctx)
				if err != nil {
					log.Println(err)
				}

				<-time.After(time.Second * 10)
			}
		}(sName, sensors)
	}
}

func (p *panoptes) unsubscribe(device config.Device) {
	cancel := p.register[device.Host]
	cancel()
}
