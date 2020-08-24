package demux

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestMQ(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nsqServer(ctx, t)

	cfg := config.NewMockConfig()

	time.Sleep(1 * time.Second)

	chMap := &extDSChanMap{eDSChan: make(map[string]telemetry.ExtDSChan)}
	testChan := make(telemetry.ExtDSChan)
	chMap.add("test", testChan)

	mq, err := NewMQ(ctx, cfg.Logger(), chMap)
	mq.batchSize = 0
	mq.drainInterval = time.Duration(10)
	assert.NoError(t, err)

	ds := telemetry.ExtDataStore{
		Output: "test::test",
		DS: telemetry.DataStore{
			"metric": "test",
		},
	}

	mq.publish(ds, "test")

	select {
	case dsQ := <-testChan:
		assert.Contains(t, dsQ.DS, "metric")
	case <-time.After(1 * time.Second):
		assert.Fail(t, "timeout")
	}
}

func TestBatchDrainer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nsqServer(ctx, t)

	cfg := config.NewMockConfig()

	time.Sleep(1 * time.Second)

	chMap := &extDSChanMap{eDSChan: make(map[string]telemetry.ExtDSChan)}
	testChan := make(telemetry.ExtDSChan)
	chMap.add("test", testChan)

	mq, err := NewMQ(ctx, cfg.Logger(), chMap)
	mq.batchSize = 10
	mq.drainInterval = time.Duration(1)
	assert.NoError(t, err)

	ds := telemetry.ExtDataStore{
		Output: "test::test",
		DS: telemetry.DataStore{
			"metric": "test",
		},
	}

	mq.publish(ds, "test")

	select {
	case dsQ := <-testChan:
		assert.Contains(t, dsQ.DS, "metric")
	case <-time.After(2 * time.Second):
		assert.Fail(t, "timeout")
	}
}

func nsqServer(ctx context.Context, t *testing.T) {
	go func() {
		cmd := exec.CommandContext(ctx, "nsqd", "-data-path", "/tmp")
		t.Log(cmd.Run())
	}()
}
