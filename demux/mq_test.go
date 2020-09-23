//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package demux

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

func TestMQ(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd, tmpDir := nsqServer(ctx, t)
	defer cmd.Process.Kill()
	defer os.RemoveAll(tmpDir)

	time.Sleep(2 * time.Second)

	t.Run("testBatchDrainer", testBatchDrainer)
	t.Run("testPublish", testPublish)
}

func testPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewMockConfig()

	chMap := &extDSChanMap{eDSChan: make(map[string]telemetry.ExtDSChan)}
	testChan := make(telemetry.ExtDSChan)
	chMap.add("test1", testChan)

	mq, err := NewMQ(ctx, cfg.Logger(), chMap)
	assert.NoError(t, err)
	mq.batchSize = 0
	mq.drainInterval = time.Duration(10)

	ds := telemetry.ExtDataStore{
		Output: "test1::test1",
		DS: telemetry.DataStore{
			"metric":    "test",
			"labels":    map[string]string{"label1": "value1"},
			"timestamp": 1599982184000000,
		},
	}

	mq.publish(ds, "test1")

	select {
	case dsQ := <-testChan:
		assert.Contains(t, dsQ.DS, "metric")
	case <-time.After(1 * time.Second):
		assert.Fail(t, "timeout")
	}
}

func testBatchDrainer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewMockConfig()

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
			"metric":    "test",
			"labels":    map[string]string{"label1": "value1"},
			"timestamp": 1599982184000000,
		},
	}

	mq.publish(ds, "test")

	select {
	case dsQ := <-testChan:
		assert.Contains(t, dsQ.DS, "metric")
	case <-time.After(5 * time.Second):
		assert.Fail(t, "timeout")
	}
}

func nsqServer(ctx context.Context, t *testing.T) (*exec.Cmd, string) {
	dir, err := ioutil.TempDir("", "nsq")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(ctx, "nsqd", "-data-path", dir, "-tcp-address", "127.0.0.1:4150", "-http-address", "127.0.0.1:4153")
	cmd.Start()

	t.Log(cmd.String())

	return cmd, dir
}
