package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

// NMIFactory ...
type NMIFactory func(*zap.Logger, *grpc.ClientConn, []*config.Sensor, ExtDSChan) NMI

// NMI ...
type NMI interface {
	Start(context.Context) error
}

// DataStore ...
type DataStore map[string]interface{}

type ExtDataStore struct {
	Output string
	DS     DataStore
}

type ExtDSChan chan ExtDataStore

type TelemetryRegistrar struct {
	nmi map[string]NMIFactory
	lg  *zap.Logger
	sync.RWMutex
}

func NewRegistrar(lg *zap.Logger) *TelemetryRegistrar {
	return &TelemetryRegistrar{
		nmi: make(map[string]NMIFactory),
		lg:  lg,
	}
}

func (tr *TelemetryRegistrar) Register(name string, tf NMIFactory) {
	tr.lg.Info("telemetry/register", zap.String("nmi", name))
	tr.set(name, tf)
}

func (tr *TelemetryRegistrar) GetNMIFactory(name string) (NMIFactory, bool) {
	return tr.get(name)
}

func (tr *TelemetryRegistrar) set(name string, nf NMIFactory) {
	tr.Lock()
	defer tr.Unlock()
	tr.nmi[name] = nf
}

func (tr *TelemetryRegistrar) get(name string) (NMIFactory, bool) {
	tr.RLock()
	defer tr.RUnlock()
	v, ok := tr.nmi[name]
	return v, ok
}

func (ds DataStore) PrettyPrint() error {
	b, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}
