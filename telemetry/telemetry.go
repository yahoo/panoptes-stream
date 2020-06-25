package telemetry

import (
	"context"
	"encoding/json"
	"fmt"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"google.golang.org/grpc"
)

// NMIFactory ...
type NMIFactory func(*grpc.ClientConn, []*config.Sensor, DSChan) NMI

// NMI ...
type NMI interface {
	Start(context.Context) error
}

// DataStore ...
type DataStore map[string]interface{}

// DSChan ...
type DSChan chan DataStore

// R ...
var R = make(map[string]NMIFactory)

// Register ...
func Register(n string, c NMIFactory) {
	R[n] = c
}

func (ds DataStore) PrettyPrint() error {
	b, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}
