package telemetry

import (
	"context"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"google.golang.org/grpc"
)

// Constructor ...
type Constructor func(*grpc.ClientConn, []*config.Sensor, KVChan) NMI

// NMI ...
type NMI interface {
	Start(context.Context) error
}

// KV ...
type KV map[string]interface{}

// KVChan ...
type KVChan chan KV

// R ...
var R = make(map[string]Constructor)

// Register ...
func Register(n string, c Constructor) {
	R[n] = c
}
