//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package consul

import (
	"os"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/assert"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

func TestNewConsul(t *testing.T) {
	srv, err := testutil.NewTestServerConfigT(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

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

	assert.Len(t, devices, 1)
	assert.Len(t, databases, 1)
	assert.Len(t, producers, 2)

	assert.Equal(t, "core1.lhr", devices[0].Host)
	assert.Equal(t, "influxdb", databases[0].Service)
	assert.Equal(t, "kafka", producers[0].Service)
	assert.Equal(t, "127.0.0.2:8081", cfg.Global().Status.Addr)
	assert.NotEqual(t, nil, cfg.Logger())

	_, ok := devices[0].Sensors["juniper.jti"]
	assert.Equal(t, true, ok)

	srv.SetKV(t, "panoptes/config/global", []byte(`{"logger": {"level":"info", "encoding": "console", "outputPaths": ["stdout"], "errorOutputPaths":["stderr"]}, "status": {"addr":"127.0.0.2:8082"}}`))
	cfg.Update()
	assert.Equal(t, "127.0.0.2:8082", cfg.Global().Status.Addr)
}

func TestGetTLSConfig(t *testing.T) {
	t.Parallel()

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

func TestEmptyConfig(t *testing.T) {
	srv, err := testutil.NewTestServerConfigT(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	os.Setenv("PANOPTES_CONFIG_CONSUL_ADDRESS", srv.HTTPAddr)

	config := map[string][]byte{"panoptes/config/": []byte("")}
	srv.PopulateKV(t, config)

	_, err = New("-")
	assert.Equal(t, nil, err)
}
