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
	"github.com/yahoo/panoptes-stream/producer"
	"github.com/yahoo/panoptes-stream/telemetry"
)

var cfg = config.NewMockConfig()

func TestConsoleStdout(t *testing.T) {
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

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

func TestConsoleStderr(t *testing.T) {
	stderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	ch := make(telemetry.ExtDSChan, 2)
	p := New(context.Background(), config.Producer{}, cfg.Logger(), ch)
	go p.Start()

	ch <- telemetry.ExtDataStore{
		Output: "console::stderr",
		DS:     telemetry.DataStore{"test": "test"},
	}

	buf := new(bytes.Buffer)
	io.CopyN(buf, r, 20)
	os.Stdout = stderr
	assert.Contains(t, buf.String(), "test")

	ch <- telemetry.ExtDataStore{
		Output: "console",
		DS:     telemetry.DataStore{"test": "test"},
	}
	time.Sleep(time.Second)
	assert.Contains(t, cfg.LogOutput.String(), "wrong output")

	close(ch)
}

func TestRegister(t *testing.T) {
	r := producer.NewRegistrar(cfg.Logger())
	Register(r)
	_, ok := r.GetProducerFactory("console")
	assert.Equal(t, true, ok)
}
