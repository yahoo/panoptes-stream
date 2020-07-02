package telemetry

import (
	"context"
	"encoding/json"
	"os"

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

func (ds DataStore) PrettyPrint(fdType string) error {
	b, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return err
	}

	if fdType == "stdout" {
		os.Stdout.Write(b)
	} else {
		os.Stderr.Write(b)
	}
	return nil
}
