//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms

package console

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/telemetry"
)

func TestConsole(t *testing.T) {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := config.NewMockConfig()
	ch := make(telemetry.ExtDSChan, 2)
	p := New(context.Background(), config.Producer{}, cfg.Logger(), ch)
	go p.Start()

	ch <- telemetry.ExtDataStore{
		Output: "console::stdout",
		DS:     telemetry.DataStore{"test": "test"},
	}

	buf := new(bytes.Buffer)
	io.CopyN(buf, r, 20)
	os.Stdout = stdout
	assert.Contains(t, buf.String(), "test")

	ch <- telemetry.ExtDataStore{
		Output: "console",
		DS:     telemetry.DataStore{"test": "test"},
	}
	time.Sleep(time.Second)
	assert.Contains(t, cfg.LogOutput.String(), "wrong output")

	close(ch)
}
