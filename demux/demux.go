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
	ctx      context.Context
	cfg      config.Config
	lg       *zap.Logger
	inChan   telemetry.ExtDSChan
	chMap    map[string]telemetry.ExtDSChan
	pr       *producer.Registrar
	register map[string]context.CancelFunc
}

func New(ctx context.Context, cfg config.Config, lg *zap.Logger, pr *producer.Registrar, inChan telemetry.ExtDSChan) *Demux {
	return &Demux{
		ctx:      ctx,
		cfg:      cfg,
		lg:       lg,
		pr:       pr,
		inChan:   inChan,
		chMap:    make(map[string]telemetry.ExtDSChan),
		register: make(map[string]context.CancelFunc),
	}
}

func (d *Demux) Init() error {
	for _, producer := range d.cfg.Producers() {
		err := d.subscribe(producer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Demux) Start() {
	for {
		extDS, _ := <-d.inChan
		output := strings.Split(extDS.Output, "::")
		if len(output) < 2 {
			continue
		}

		d.chMap[output[0]] <- extDS

	}
}

func (d *Demux) subscribe(producer config.Producer) error {
	var ctx context.Context

	mqNew, ok := d.pr.GetProducerFactory(producer.Service)
	if !ok {
		return errors.New("producer not exist")
	}

	// register channel
	d.chMap[producer.Name] = make(telemetry.ExtDSChan, 1)
	// register cancelFunnc
	ctx, d.register[producer.Name] = context.WithCancel(d.ctx)
	// construct
	m := mqNew(ctx, producer, d.lg, d.chMap[producer.Name])
	// start the producer
	go m.Start()

	return nil
}

func (d *Demux) unsubscribe(producer config.Producer) {
	cancel := d.register[producer.Name]
	cancel()
}
