package telemetry

import (
	"strings"
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
	// name convention: service[::ext%d] example: cisco.gnmi or cisco.gnmi::ext1
	service := strings.Split(name, "::")
	if len(service) < 1 {
		return nil, false
	}

	return tr.get(service[0])
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
