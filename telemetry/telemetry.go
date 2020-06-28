package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	reg = &telemetryRegister{nmi: make(map[string]NMIFactory)}
	lg  = zap.NewNop()
)

type telemetryRegister struct {
	nmi map[string]NMIFactory
	sync.RWMutex
}

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

func (tr *telemetryRegister) set(name string, nf NMIFactory) {
	tr.Lock()
	defer tr.Unlock()
	tr.nmi[name] = nf
}

func (tr *telemetryRegister) get(name string) NMIFactory {
	tr.RLock()
	defer tr.RUnlock()
	return tr.nmi[name]
}

// Register ...
func Register(n string, nf NMIFactory) {
	lg.Info("telemetry/register", zap.String("nmi", n))
	reg.set(n, nf)
}

func GetNMIFactory(name string) NMIFactory {
	return reg.get(name)
}

func (ds DataStore) PrettyPrint() error {
	b, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}

func SetLogger(logger *zap.Logger) {
	lg = logger
}
