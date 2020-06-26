package producer

import (
	"sync"
)

var reg = producerRegister{p: make(map[string]Producer)}

type producerRegister struct {
	p map[string]Producer
	sync.RWMutex
}

func (pr *producerRegister) set(name string, m Producer) {
	pr.Lock()
	defer pr.Unlock()
	pr.p[name] = m
}

func (pr *producerRegister) get(name string) (Producer, bool) {
	pr.RLock()
	defer pr.RUnlock()
	v, ok := pr.p[name]

	return v, ok
}

type Producer interface {
	Setup()
	Start()
}

func Register(name string, p Producer) {
	reg.set(name, p)
}

func New(name string) (Producer, bool) {
	p, ok := reg.get(name)
	return p, ok
}
