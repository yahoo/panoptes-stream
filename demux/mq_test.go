package demux

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestMQ(t *testing.T) {
	os.Setenv("PANOPTES_NSQ_ADDR", "127.0.0.1:4155")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nsqServer(ctx, t)

	cfg := config.NewMockConfig()

	time.Sleep(1 * time.Second)

	chMap := &extDSChanMap{eDSChan: make(map[string]telemetry.ExtDSChan)}
	testChan := make(telemetry.ExtDSChan)
	chMap.add("test", testChan)

	mq, err := NewMQ(ctx, cfg.Logger(), chMap)
	assert.NoError(t, err)
	mq.batchSize = 0
	mq.drainInterval = time.Duration(10)

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
	os.Setenv("PANOPTES_NSQ_ADDR", "127.0.0.1:4155")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	nsqServer(ctx, t)

	cfg := config.NewMockConfig()

	time.Sleep(1 * time.Second)

	chMap := &extDSChanMap{eDSChan: make(map[string]telemetry.ExtDSChan)}
	testChan := make(telemetry.ExtDSChan)
	chMap.add("test", testChan)

	mq, err := NewMQ(ctx, cfg.Logger(), chMap)
	assert.NoError(t, err)
	mq.batchSize = 10
	mq.drainInterval = time.Duration(1)

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
		addr := os.Getenv("PANOPTES_NSQ_ADDR")
		cmd := exec.CommandContext(ctx, "nsqd", "-data-path", "/tmp", "-tcp-address", addr)
		t.Log(cmd.Run())
	}()

	time.Sleep(200 * time.Millisecond)
}
