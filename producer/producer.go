package producer

import (
	"sync"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

var reg = producerRegister{p: make(map[string]ProducerFactory)}

type ProducerFactory func(config.Producer, telemetry.ExtDSChan) Producer

type producerRegister struct {
	p map[string]ProducerFactory
	sync.RWMutex
}

func (pr *producerRegister) set(name string, m ProducerFactory) {
	pr.Lock()
	defer pr.Unlock()
	pr.p[name] = m
}

func (pr *producerRegister) get(name string) (ProducerFactory, bool) {
	pr.RLock()
	defer pr.RUnlock()
	v, ok := pr.p[name]

	return v, ok
}

type Producer interface {
	Start()
}

func Register(name string, pf ProducerFactory) {
	reg.set(name, pf)
}

func GetProducerFactory(name string) (ProducerFactory, bool) {
	return reg.get(name)
}
