//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package influxdb

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

func TestLineProtocol(t *testing.T) {
	data := telemetry.ExtDataStore{
		Output: "influx1::ifcounters",
		DS: telemetry.DataStore{
			"key":       "out-octets",
			"labels":    map[string]string{"name": "Ethernet3"},
			"prefix":    "/interfaces/interface/state/counters/",
			"system_id": "core1.bur",
			"timestamp": 1595768623436661269,
			"value":     5587651,
		},
	}

	buf := new(bytes.Buffer)

	l, err := getLineProtocol(buf, data)
	require.Equal(t, err, nil)
	assert.Equal(t, l, "ifcounters,_prefix_=/interfaces/interface/state/counters/,_host_=core1.bur,name=Ethernet3 out-octets=5587651 1595768623436661269")
}

func TestSingleMetric(t *testing.T) {
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		assert.Equal(t, "test,_prefix_=/tests/test,_host_=127.0.0.1 mykey=0 150000000\n", string(body))
		w.WriteHeader(http.StatusAccepted)
		close(done)
	}))

	cfg := config.NewMockConfig()
	ch := make(telemetry.ExtDSChan, 10)

	dbCfg := config.Database{Name: "influxdb1", Service: "influxdb", Config: map[string]interface{}{
		"server":     server.URL,
		"bucket":     "mybucket",
		"bactchSize": 1,
	}}

	db := New(ctx, dbCfg, cfg.Logger(), ch)
	go db.Start()
	ch <- telemetry.ExtDataStore{
		Output: "influxdb1::test",
		DS: map[string]interface{}{
			"prefix":    "/tests/test",
			"labels":    map[string]string{},
			"system_id": "127.0.0.1",
			"timestamp": 150000000,
			"key":       "mykey",
			"value":     0,
		},
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Error("time limit exceeded")
	}

	time.Sleep(time.Second)
	cancel()
}

func BenchmarkLineProtocol(b *testing.B) {
	data := telemetry.ExtDataStore{
		Output: "influx1::ifcounters",
		DS: telemetry.DataStore{
			"key":       "out-octets",
			"labels":    map[string]string{"name": "Ethernet3"},
			"prefix":    "/interfaces/interface/state/counters/",
			"system_id": "core1.bur",
			"timestamp": 1595768623436661269,
			"value":     5587651,
		},
	}

	buf := new(bytes.Buffer)

	for i := 0; i < b.N; i++ {
		getLineProtocol(buf, data)
	}
}

func TestGetValueString(t *testing.T) {
	var v interface{}
	v = 1
	assert.Equal(t, "1", getValueString(v))
	v = 1.5
	assert.Equal(t, "1.500000", getValueString(v))
	v = true
	assert.Equal(t, "true", getValueString(v))
	v = "test"
	assert.Equal(t, "\"test\"", getValueString(v))
	v = []byte("test")
	assert.Equal(t, "\"test\"", getValueString(v))
	v = map[string]interface{}{"name": "foo"}
	assert.Equal(t, "\"{\\\"name\\\":\\\"foo\\\"}\"", getValueString(v))
	v = map[string]string{"a": "b"}
	assert.Equal(t, "", getValueString(v))
}
