package demux

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/database"
	"git.vzbuilders.com/marshadrad/panoptes/producer"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

// Demux represents demux
type Demux struct {
	ctx       context.Context
	cfg       config.Config
	logger    *zap.Logger
	inChan    telemetry.ExtDSChan
	chMap     *extDSChanMap
	pr        *producer.Registrar
	db        *database.Registrar
	mq        *MQ
	register  map[string]context.CancelFunc
	producers map[string]config.Producer
	databases map[string]config.Database
}

type extDSChanMap struct {
	sync.RWMutex
	eDSChan map[string]telemetry.ExtDSChan
}

// New constructs new demux
func New(ctx context.Context, cfg config.Config, pr *producer.Registrar, db *database.Registrar, inChan telemetry.ExtDSChan) *Demux {
	return &Demux{
		ctx:       ctx,
		cfg:       cfg,
		logger:    cfg.Logger(),
		pr:        pr,
		db:        db,
		inChan:    inChan,
		chMap:     &extDSChanMap{eDSChan: make(map[string]telemetry.ExtDSChan)},
		register:  make(map[string]context.CancelFunc),
		producers: make(map[string]config.Producer),
		databases: make(map[string]config.Database),
	}
}

func (d *Demux) init() error {
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

// Start starts demux
func (d *Demux) Start() {
	d.init()

	d.mq, _ = NewMQ(d.ctx, d.logger, d.chMap)
	if d.mq != nil {
		d.logger.Info("demux", zap.String("event", "NSQ enabled"))
	}

	go func() {
		d.start()
	}()
}

func (d *Demux) start() {
	var (
		outChan telemetry.ExtDSChan
		ok      bool
	)

	for {
		extDS := <-d.inChan

		output := strings.Split(extDS.Output, "::")
		if len(output) < 2 {
			d.logger.Error("demux", zap.String("error", "output not found"))
			continue
		}

		if outChan, ok = d.chMap.get(output[0]); !ok {
			d.logger.Error("demux", zap.String("error", "channel not found"), zap.String("name", output[0]))
			continue
		}

		select {
		case outChan <- extDS:

		case <-d.ctx.Done():
			d.logger.Info("demux has been terminated")
			return

		default:
			if d.mq != nil {
				d.mq.publish(extDS, output[0])
				continue
			}

			d.logger.Warn("demux", zap.String("error", "dataset drop"), zap.String("name", output[0]))
		}
	}
}

func (d *Demux) subscribeProducer(producer config.Producer) error {
	var ctx context.Context

	new, ok := d.pr.GetProducerFactory(producer.Service)
	if !ok {
		return errors.New("producer not exist")
	}

	// register producer
	d.producers[producer.Name] = producer
	// make channel
	ch := make(telemetry.ExtDSChan, d.cfg.Global().OutputBufferSize)
	// register channel
	d.chMap.add(producer.Name, ch)
	// register cancelFunnc
	ctx, d.register[producer.Name] = context.WithCancel(d.ctx)
	// construct
	p := new(ctx, producer, d.logger, ch)
	// start the producer
	go p.Start()

	return nil
}

func (d *Demux) subscribeDatabase(database config.Database) error {
	var ctx context.Context

	new, ok := d.db.GetDatabaseFactory(database.Service)
	if !ok {
		d.logger.Info(database.Service)
		return errors.New("database not exist")
	}

	// register database
	d.databases[database.Name] = database
	// make a channel
	ch := make(telemetry.ExtDSChan, d.cfg.Global().OutputBufferSize)
	// register channel
	d.chMap.add(database.Name, ch)
	// register cancelFunnc
	ctx, d.register[database.Name] = context.WithCancel(d.ctx)
	// construct
	db := new(ctx, database, d.logger, ch)
	// start the database agent
	go db.Start()

	return nil
}

func (d *Demux) unsubscribeProducer(producer config.Producer) {
	d.register[producer.Name]()
	delete(d.producers, producer.Name)
	delete(d.register, producer.Name)
	d.chMap.del(producer.Name)
}

func (d *Demux) unsubscribeDatabase(database config.Database) {
	d.register[database.Name]()
	delete(d.databases, database.Name)
	delete(d.register, database.Name)
	d.chMap.del(database.Name)
}

// Update updates databases and producers
func (d *Demux) Update() {
	d.updateProducer()
	d.updateDatabase()

	if d.mq != nil {
		d.mq.update()
	}
}

func (d *Demux) updateDatabase() {
	newDatabases := make(map[string]config.Database)
	delta := &struct {
		add []config.Database
		del []config.Database
		mod []config.Database
	}{}

	for _, database := range d.cfg.Databases() {
		newDatabases[database.Name] = database

		if _, ok := d.databases[database.Name]; !ok {
			delta.add = append(delta.add, database)
			continue
		}

		if ok := reflect.DeepEqual(d.databases[database.Name], database); !ok {
			delta.mod = append(delta.mod, database)
		}
	}

	for name, database := range d.databases {
		if _, ok := newDatabases[name]; !ok {
			delta.del = append(delta.del, database)
		}
	}

	for _, database := range delta.add {
		d.subscribeDatabase(database)
	}

	for _, database := range delta.del {
		d.unsubscribeDatabase(database)
	}

	for _, database := range delta.mod {
		d.unsubscribeDatabase(database)
		d.subscribeDatabase(database)
	}
}

func (d *Demux) updateProducer() {
	newProducers := make(map[string]config.Producer)
	delta := &struct {
		add []config.Producer
		del []config.Producer
		mod []config.Producer
	}{}

	for _, producer := range d.cfg.Producers() {
		newProducers[producer.Name] = producer

		if _, ok := d.producers[producer.Name]; !ok {
			delta.add = append(delta.add, producer)
			continue
		}

		if ok := reflect.DeepEqual(d.producers[producer.Name], producer); !ok {
			delta.mod = append(delta.mod, producer)
		}
	}

	for name, producer := range d.producers {
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

func (e *extDSChanMap) get(key string) (telemetry.ExtDSChan, bool) {
	e.RLock()
	defer e.RUnlock()
	v, ok := e.eDSChan[key]
	return v, ok
}

func (e *extDSChanMap) add(key string, value telemetry.ExtDSChan) {
	e.Lock()
	defer e.Unlock()
	e.eDSChan[key] = value
}

func (e *extDSChanMap) del(key string) {
	e.Lock()
	defer e.Unlock()
	delete(e.eDSChan, key)
}

func (e *extDSChanMap) list() []string {
	var r []string
	e.RLock()
	defer e.RUnlock()
	for k := range e.eDSChan {
		r = append(r, k)
	}

	return r
}
