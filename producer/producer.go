package producer

import (
	"sync"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"go.uber.org/zap"
)

type ProducerFactory func(config.Producer, *zap.Logger, telemetry.ExtDSChan) Producer

type Producer interface {
	Start()
}

type ProducerRegistrar struct {
	p  map[string]ProducerFactory
	lg *zap.Logger
	sync.RWMutex
}

func NewRegistrar(lg *zap.Logger) *ProducerRegistrar {
	return &ProducerRegistrar{
		p:  make(map[string]ProducerFactory),
		lg: lg,
	}
}

func (pr *ProducerRegistrar) Register(name string, pf ProducerFactory) {
	pr.lg.Info("producer/register", zap.String("mq", name))
	pr.set(name, pf)
}

func (pr *ProducerRegistrar) GetProducerFactory(name string) (ProducerFactory, bool) {
	return pr.get(name)
}

func (pr *ProducerRegistrar) set(name string, m ProducerFactory) {
	pr.Lock()
	defer pr.Unlock()
	pr.p[name] = m
}

func (pr *ProducerRegistrar) get(name string) (ProducerFactory, bool) {
	pr.RLock()
	defer pr.RUnlock()
	v, ok := pr.p[name]

	return v, ok
}
