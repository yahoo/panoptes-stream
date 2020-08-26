package producer

import (
	"sync"

	"go.uber.org/zap"
)

// Registrar represents producer factory registration.
type Registrar struct {
	p      map[string]Factory
	logger *zap.Logger
	sync.RWMutex
}

// NewRegistrar creates new registrar.
func NewRegistrar(logger *zap.Logger) *Registrar {
	return &Registrar{
		p:      make(map[string]Factory),
		logger: logger,
	}
}

// Register adds new producer factory
func (pr *Registrar) Register(name, vendor string, pf Factory) {
	pr.logger.Info("producer/register", zap.String("name", name), zap.String("vendor", vendor))
	pr.set(name, pf)
}

// GetProducerFactory returns requested producer factory.
func (pr *Registrar) GetProducerFactory(name string) (Factory, bool) {
	return pr.get(name)
}

// set registers a producer factory.
func (pr *Registrar) set(name string, m Factory) {
	pr.Lock()
	defer pr.Unlock()
	pr.p[name] = m
}

// get returns requested producer factory.
func (pr *Registrar) get(name string) (Factory, bool) {
	pr.RLock()
	defer pr.RUnlock()
	v, ok := pr.p[name]

	return v, ok
}
