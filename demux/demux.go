package demux

import (
	"context"
	"errors"
	"strings"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

type Demux struct {
	cfg    config.Config
	inChan telemetry.ExtDSChan
	chMap  map[string]telemetry.ExtDSChan
}

func New(cfg config.Config, inChan telemetry.ExtDSChan) *Demux {
	return &Demux{
		cfg:    cfg,
		inChan: inChan,
		chMap:  make(map[string]telemetry.ExtDSChan),
	}
}

func (d *Demux) Init() error {
	// TODO same proc for db
	for _, p := range d.cfg.Producers() {
		mqNew, ok := producer.GetProducerFactory(p.Service)
		if !ok {
			return errors.New("producer not exist")
		}

		// register channel
		d.chMap[p.Name] = make(telemetry.ExtDSChan, 1)
		// construct
		m := mqNew(p, d.chMap[p.Name])
		// start the producer
		go m.Start()
	}

	return nil
}

func (d *Demux) Start(ctx context.Context) {
	for {
		extDS, _ := <-d.inChan
		output := strings.Split(extDS.Output, "::")
		if len(output) < 2 {
			continue
		}

		d.chMap[output[0]] <- extDS

	}
}
