//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package consul

import (
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/yahoo/panoptes-stream/config"
)

var srv *testutil.TestServer

func TestConsul(t *testing.T) {
	var err error
	srv, err = testutil.NewTestServerConfigT(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	time.Sleep(time.Second)

	t.Run("testNewConsul", testNewConsul)
	t.Run("testEmptyConfig", testEmptyConfig)
	t.Run("testSignalHandler", testSignalHandler)
}

func testNewConsul(t *testing.T) {
	config := map[string][]byte{
		"panoptes/config/devices/core1.bur": []byte(`{"host": "core1.lhr", "port": 50051,  "sensors" : ["sensor1"]}`),
		"panoptes/config/sensors/sensor1":   []byte(`{"service": "juniper.jti", "path": "/interfaces/", "mode": "sample", "sampleInterval": 10, "output":"console::stdout"}`),
		"panoptes/config/databases/db1":     []byte(`{"service": "influxdb", "config": {"server": "https://localhost:8086"}}`),
		"panoptes/config/producers/kafka1":  []byte(`{"service": "kafka", "config" : {"brokers": ["127.0.0.1:9092"], "topics":["bgp"]}}`),
		"panoptes/config/global":            []byte(`{"logger": {"level":"info", "encoding": "console", "outputPaths": ["stdout"], "errorOutputPaths":["stderr"]}, "status": {"addr":"127.0.0.2:8081"}}`),
	}

	srv.PopulateKV(t, config)

	os.Setenv("PANOPTES_CONFIG_CONSUL_ADDRESS", srv.HTTPAddr)

	cfg, err := New("-")
	assert.Equal(t, nil, err)

	devices := cfg.Devices()
	databases := cfg.Databases()
	producers := cfg.Producers()
	sensors := cfg.Sensors()

	assert.Len(t, devices, 1)
	assert.Len(t, databases, 1)
	assert.Len(t, producers, 2)
	assert.Len(t, sensors, 1)

	assert.Equal(t, "core1.lhr", devices[0].Host)
	assert.Equal(t, "influxdb", databases[0].Service)
	assert.Equal(t, "kafka", producers[0].Service)
	assert.Equal(t, "juniper.jti", sensors[0].Service)
	assert.NotNil(t, cfg.Informer())
	assert.Equal(t, "127.0.0.2:8081", cfg.Global().Status.Addr)
	assert.NotEqual(t, nil, cfg.Logger())

	_, ok := devices[0].Sensors["juniper.jti"]
	assert.Equal(t, true, ok)

	srv.SetKV(t, "panoptes/config/global", []byte(`{"logger": {"level":"info", "encoding": "console", "outputPaths": ["stdout"], "errorOutputPaths":["stderr"]}, "status": {"addr":"127.0.0.2:8082"}}`))
	cfg.Update()
	assert.Equal(t, "127.0.0.2:8082", cfg.Global().Status.Addr)
}

func TestGetTLSConfig(t *testing.T) {
	cfg := &consulConfig{}
	tls, err := getTLSConfig(cfg)
	assert.Equal(t, nil, err)
	assert.Equal(t, api.TLSConfig{}, tls)

	cfg = &consulConfig{
		TLSConfig: config.TLSConfig{
			CertFile:           "/etc/certs/cert.file",
			InsecureSkipVerify: true,
		},
	}
	tls, err = getTLSConfig(cfg)
	assert.Equal(t, nil, err)
	assert.Equal(t, "/etc/certs/cert.file", tls.CertFile)
	assert.Equal(t, true, tls.InsecureSkipVerify)
}

func testEmptyConfig(t *testing.T) {
	os.Setenv("PANOPTES_CONFIG_CONSUL_ADDRESS", srv.HTTPAddr)

	config := map[string][]byte{"panoptes/config/": []byte("")}
	srv.PopulateKV(t, config)

	_, err := New("-")
	assert.Equal(t, nil, err)
}

func testSignalHandler(t *testing.T) {
	ch := make(chan struct{}, 1)
	cfg := config.NewMockConfig()
	c := &consul{
		informer: ch,
		logger:   cfg.Logger(),
	}

	go c.signalHandler()
	time.Sleep(time.Second)

	proc, err := os.FindProcess(os.Getpid())
	assert.NoError(t, err)
	proc.Signal(syscall.SIGHUP)

	select {
	case <-ch:
	case <-time.After(time.Second):
		assert.Fail(t, "time exceeded")
	}

	proc.Signal(syscall.SIGHUP)
	time.Sleep(100 * time.Millisecond)
	proc.Signal(syscall.SIGHUP)

	time.Sleep(time.Second)
	assert.Contains(t, cfg.LogOutput.String(), "dropped")
}

func TestMix(t *testing.T) {
	ch := make(chan struct{}, 1)
	c := &consul{
		devices:   []config.Device{},
		databases: []config.Database{},
		producers: []config.Producer{},
		sensors:   []config.Sensor{},
		global:    &config.Global{BufferSize: 5},
		informer:  ch,
		logger:    config.GetDefaultLogger(),
	}

	g := c.Global()
	assert.Equal(t, 5, g.BufferSize)
	i := c.Informer()
	assert.NotNil(t, i)
	d := c.Databases()
	assert.NotNil(t, d)
	p := c.Producers()
	assert.NotNil(t, p)
	s := c.Sensors()
	assert.NotNil(t, s)
	dd := c.Devices()
	assert.NotNil(t, dd)
	l := c.Logger()
	assert.NotNil(t, l)
}
