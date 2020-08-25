package database

import (
	"sync"

	"go.uber.org/zap"
)

// Registrar represents database's factories
type Registrar struct {
	p  map[string]DatabaseFactory
	lg *zap.Logger
	sync.RWMutex
}

// NewRegistrar creates new registrar.
func NewRegistrar(lg *zap.Logger) *Registrar {
	return &Registrar{
		p:  make(map[string]DatabaseFactory),
		lg: lg,
	}
}

// Register adds new database factory.
func (r *Registrar) Register(name, vendor string, df DatabaseFactory) {
	r.lg.Info("database/register", zap.String("name", name), zap.String("vendor", vendor))
	r.set(name, df)
}

// GetDatabaseFactory returns requested database factory
func (r *Registrar) GetDatabaseFactory(name string) (DatabaseFactory, bool) {
	return r.get(name)
}

// set registers a database factory
func (r *Registrar) set(name string, m DatabaseFactory) {
	r.Lock()
	defer r.Unlock()
	r.p[name] = m
}

// get returns requested database factory
func (r *Registrar) get(name string) (DatabaseFactory, bool) {
	r.RLock()
	defer r.RUnlock()
	v, ok := r.p[name]

	return v, ok
}
