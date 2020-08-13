package telemetry

import (
	"context"
	"testing"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type testNMI struct{}

func (testNMI) Start(ctx context.Context) error { return nil }

func NewNMI(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan ExtDSChan) NMI {
	return testNMI{}
}

func TestRegister(t *testing.T) {
	logger := config.GetDefaultLogger()
	registerar := NewRegistrar(logger)
	registerar.Register("test", "0.0.0", NewNMI)
	_, ok := registerar.GetNMIFactory("test")

	assert.Equal(t, true, ok)
}
