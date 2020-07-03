package demux

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/database"
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
	db       *database.Registrar
	register map[string]context.CancelFunc
}

type delta struct {
	add []config.Producer
	del []config.Producer
	mod []config.Producer
}

func New(ctx context.Context, cfg config.Config, lg *zap.Logger, pr *producer.Registrar,
	db *database.Registrar, inChan telemetry.ExtDSChan) *Demux {
	return &Demux{
		ctx:      ctx,
		cfg:      cfg,
		lg:       lg,
		pr:       pr,
		db:       db,
		inChan:   inChan,
		chMap:    make(map[string]telemetry.ExtDSChan),
		register: make(map[string]context.CancelFunc),
	}
}

func (d *Demux) Init() error {
	// producer
	for _, producer := range d.cfg.Producers() {
		err := d.subscribeProducer(producer)
		if err != nil {
			return err
		}
	}

	// database
	for _, database := range d.cfg.Databases() {
		err := d.subscribeDatabase(database)
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

		if _, ok := d.chMap[output[0]]; ok {
			d.chMap[output[0]] <- extDS
		} else {
			d.lg.Error("channel not found", zap.String("name", output[0]))
		}
	}
}

func (d *Demux) subscribeProducer(producer config.Producer) error {
	var ctx context.Context

	New, ok := d.pr.GetProducerFactory(producer.Service)
	if !ok {
		return errors.New("producer not exist")
	}

	// register channel
	d.chMap[producer.Name] = make(telemetry.ExtDSChan, 1)
	// register cancelFunnc
	ctx, d.register[producer.Name] = context.WithCancel(d.ctx)
	// construct
	m := New(ctx, producer, d.lg, d.chMap[producer.Name])
	// start the producer
	go m.Start()

	return nil
}

func (d *Demux) subscribeDatabase(database config.Database) error {
	var ctx context.Context

	New, ok := d.db.GetDatabaseFactory(database.Service)
	if !ok {
		d.lg.Info(database.Service)
		return errors.New("database not exist")
	}

	// register channel
	d.chMap[database.Name] = make(telemetry.ExtDSChan, 1)
	// register cancelFunnc
	ctx, d.register[database.Name] = context.WithCancel(d.ctx)
	// construct
	db := New(ctx, database, d.lg, d.chMap[database.Name])
	// start the database agent
	go db.Start()

	return nil
}

func (d *Demux) unsubscribeProducer(producer config.Producer) {
	d.register[producer.Name]()
	delete(d.register, producer.Name)
	delete(d.chMap, producer.Name)
}

func (d *Demux) Update(producers map[string]config.Producer) {
	newProducers := make(map[string]config.Producer)
	delta := new(delta)

	for _, producer := range d.cfg.Producers() {
		newProducers[producer.Name] = producer

		if _, ok := producers[producer.Name]; !ok {
			delta.add = append(delta.add, producer)
		} else {
			if ok := reflect.DeepEqual(producers[producer.Name], producer); !ok {
				delta.mod = append(delta.mod, producer)
			}
		}

	}

	for name, producer := range producers {
		if _, ok := newProducers[name]; !ok {
			delta.del = append(delta.del, producer)
		}
	}

	for _, producer := range delta.add {
		d.subscribeProducer(producer)
	}

	for _, producer := range delta.del {
		d.unsubscribeProducer(producer)
	}

	for _, producer := range delta.mod {
		d.unsubscribeProducer(producer)
		d.subscribeProducer(producer)
	}
}
