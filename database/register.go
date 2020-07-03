package database

import (
	"sync"

	"go.uber.org/zap"
)

type Registrar struct {
	p  map[string]DatabaseFactory
	lg *zap.Logger
	sync.RWMutex
}

func NewRegistrar(lg *zap.Logger) *Registrar {
	return &Registrar{
		p:  make(map[string]DatabaseFactory),
		lg: lg,
	}
}

func (r *Registrar) Register(name, vendor string, df DatabaseFactory) {
	r.lg.Info("database/register", zap.String("name", name), zap.String("vendor", vendor))
	r.set(name, df)
}

func (r *Registrar) GetDatabaseFactory(name string) (DatabaseFactory, bool) {
	return r.get(name)
}

func (r *Registrar) set(name string, m DatabaseFactory) {
	r.Lock()
	defer r.Unlock()
	r.p[name] = m
}

func (r *Registrar) get(name string) (DatabaseFactory, bool) {
	r.RLock()
	defer r.RUnlock()
	v, ok := r.p[name]

	return v, ok
}
