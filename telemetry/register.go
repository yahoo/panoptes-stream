package telemetry

import (
	"sync"

	"go.uber.org/zap"
)

type Registrar struct {
	nmi map[string]NMIFactory
	lg  *zap.Logger
	sync.RWMutex
}

func NewRegistrar(lg *zap.Logger) *Registrar {
	return &Registrar{
		nmi: make(map[string]NMIFactory),
		lg:  lg,
	}
}

func (tr *Registrar) Register(name, version string, tf NMIFactory) {
	tr.lg.Info("telemetry/register", zap.String("name", name), zap.String("version", version))
	tr.set(name, tf)
}

func (tr *Registrar) GetNMIFactory(name string) (NMIFactory, bool) {
	return tr.get(name)
}

func (tr *Registrar) set(name string, nf NMIFactory) {
	tr.Lock()
	defer tr.Unlock()
	tr.nmi[name] = nf
}

func (tr *Registrar) get(name string) (NMIFactory, bool) {
	tr.RLock()
	defer tr.RUnlock()
	v, ok := tr.nmi[name]
	return v, ok
}
