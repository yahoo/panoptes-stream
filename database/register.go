//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package database

import (
	"sync"

	"go.uber.org/zap"
)

// Registrar represents database factory registration.
type Registrar struct {
	p  map[string]Factory
	lg *zap.Logger
	sync.RWMutex
}

// NewRegistrar creates new registrar.
func NewRegistrar(lg *zap.Logger) *Registrar {
	return &Registrar{
		p:  make(map[string]Factory),
		lg: lg,
	}
}

// Register adds new database factory.
func (r *Registrar) Register(name, vendor string, df Factory) {
	r.lg.Info("database", zap.String("event", "register"), zap.String("name", name), zap.String("vendor", vendor))
	r.set(name, df)
}

// GetDatabaseFactory returns requested database factory.
func (r *Registrar) GetDatabaseFactory(name string) (Factory, bool) {
	return r.get(name)
}

// set registers a database factory.
func (r *Registrar) set(name string, m Factory) {
	r.Lock()
	defer r.Unlock()
	r.p[name] = m
}

// get returns requested database factory.
func (r *Registrar) get(name string) (Factory, bool) {
	r.RLock()
	defer r.RUnlock()
	v, ok := r.p[name]

	return v, ok
}
