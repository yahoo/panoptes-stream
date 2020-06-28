package demux

import (
	"context"
	"errors"
	"strings"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

type Demux struct {
	cfg    config.Config
	lg     *zap.Logger
	inChan telemetry.ExtDSChan
	chMap  map[string]telemetry.ExtDSChan
	pr     *producer.ProducerRegistrar
}

func New(cfg config.Config, lg *zap.Logger, pr *producer.ProducerRegistrar, inChan telemetry.ExtDSChan) *Demux {
	return &Demux{
		cfg:    cfg,
		lg:     lg,
		inChan: inChan,
		chMap:  make(map[string]telemetry.ExtDSChan),
		pr:     pr,
	}
}

func (d *Demux) Init() error {
	// TODO same proc for db
	for _, p := range d.cfg.Producers() {
		mqNew, ok := d.pr.GetProducerFactory(p.Service)
		if !ok {
			return errors.New("producer not exist")
		}

		// register channel
		d.chMap[p.Name] = make(telemetry.ExtDSChan, 1)
		// construct
		m := mqNew(p, d.lg, d.chMap[p.Name])
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
