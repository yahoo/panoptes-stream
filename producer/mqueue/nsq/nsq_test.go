//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package nsq

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"github.com/nsqio/go-nsq"
	"github.com/stretchr/testify/assert"
)

type messageHandler struct {
	ch  chan map[string]string
	err error
}

func TestStart(t *testing.T) {
	mCfg := config.NewMockConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, tmpDir := nsqServer(ctx, t)
	defer os.RemoveAll(tmpDir)

	time.Sleep(time.Second)

	ch := make(telemetry.ExtDSChan, 1)
	cfg := config.Producer{
		Name:    "nsq01",
		Service: "nsq",
		Config: map[string]interface{}{
			"addr":         "127.0.0.1:4165",
			"topics":       []string{"bgp", "interface"},
			"batchSize":    1,
			"batchTimeout": 1,
		},
	}

	p := New(ctx, cfg, mCfg.Logger(), ch)
	go p.Start()

	ch <- telemetry.ExtDataStore{
		Output: "nsq01::bgp",
		DS: map[string]interface{}{
			"test": "test",
		},
	}

	time.Sleep(2 * time.Second)

	// consumer
	config := nsq.NewConfig()
	consumer, err := nsq.NewConsumer("bgp", "channel", config)
	assert.NoError(t, err)
	consumer.SetLogger(&noLogger{}, 0)

	chout := make(chan map[string]string, 1)
	handler := &messageHandler{
		ch: chout,
	}

	consumer.AddConcurrentHandlers(handler, 1)

	err = consumer.ConnectToNSQD("127.0.0.1:4165")
	assert.NoError(t, err)

	t.Log(mCfg.LogOutput.String())

	select {
	case v := <-chout:
		assert.Equal(t, "test", v["test"])
	}
}

func (h *messageHandler) HandleMessage(m *nsq.Message) error {
	var ds map[string]string
	json.Unmarshal(m.Body, &ds)

	select {
	case h.ch <- ds:
	default:
		return h.err
	}

	return nil
}

func nsqServer(ctx context.Context, t *testing.T) (*exec.Cmd, string) {
	dir, err := ioutil.TempDir("", "nsq")
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.CommandContext(ctx, "nsqd", "-data-path", dir, "-tcp-address", "127.0.0.1:4165")
	cmd.Start()

	t.Log(cmd.String())

	return cmd, dir
}
