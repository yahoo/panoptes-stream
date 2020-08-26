package telemetry

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

// NMIFactory is a function that returns a new instance of a NMI
type NMIFactory func(*zap.Logger, *grpc.ClientConn, []*config.Sensor, ExtDSChan) NMI

// NMI represents a NMI
type NMI interface {
	Start(context.Context) error
}

// DataStore represents a metric and its meta data
// Meta data:
// - system_id
// - labels
// - timestamp
// - prefix
type DataStore map[string]interface{}

// ExtDataStore represents datastore with output identification
type ExtDataStore struct {
	Output string
	DS     DataStore
}

// ExtDSChan represents ExtDataStore channel
type ExtDSChan chan ExtDataStore
