//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package status

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/yahoo/panoptes-stream/config"
)

func TestHealthcheck(t *testing.T) {
	ts := httptest.NewServer(new(healthcheck))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.Equal(t, nil, err)

	hcMsg, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Equal(t, nil, err)

	assert.Equal(t, "panoptes alive and reachable", string(hcMsg))
}

func TestDuplicateRegisterMetrics(t *testing.T) {
	var (
		metrics         = make(map[string]Metrics)
		duplicateRegErr prometheus.AlreadyRegisteredError
	)

	metrics["errorsTotal"] = NewCounter("test_errors_total", "")
	metrics["processNSecond"] = NewGauge("test_process_nanosecond", "")

	Register(Labels{"host": "127.0.0.1"}, metrics)

	err := prometheus.Register(prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name:        "panoptes_test_errors_total",
		Help:        "",
		ConstLabels: map[string]string{"host": "127.0.0.1"},
	}, func() float64 { return float64(0) }))

	if !errors.As(err, &duplicateRegErr) {
		assert.Fail(t, "duplicate registration failed")
	}

	err = prometheus.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "panoptes_test_process_nanosecond",
		Help:        "",
		ConstLabels: map[string]string{"host": "127.0.0.1"},
	}, func() float64 { return float64(0) }))

	if !errors.As(err, &duplicateRegErr) {
		assert.Fail(t, "duplicate registration failed")
	}

	Unregister(Labels{"host": "127.0.0.1"}, metrics)

	err = prometheus.Register(prometheus.NewCounterFunc(prometheus.CounterOpts{
		Name:        "panoptes_test_errors_total",
		Help:        "",
		ConstLabels: map[string]string{"host": "127.0.0.1"},
	}, func() float64 { return float64(0) }))

	if errors.As(err, &duplicateRegErr) {
		assert.Fail(t, "duplicate registration failed")
	}

	err = prometheus.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name:        "panoptes_test_process_nanosecond",
		Help:        "",
		ConstLabels: map[string]string{"host": "127.0.0.1"},
	}, func() float64 { return float64(0) }))

	if errors.As(err, &duplicateRegErr) {
		assert.Fail(t, "duplicate registration failed")
	}
}

func TestMetricsFunc(t *testing.T) {
	mg := NewGauge("test_process_nanosecond", "")
	assert.Equal(t, uint64(0), mg.Value)
	mg.Inc()
	assert.Equal(t, uint64(1), mg.Value)
	mg.Dec()
	assert.Equal(t, uint64(0), mg.Value)
	mg.Set(55)
	assert.Equal(t, uint64(55), mg.Get())

	mc := NewCounter("test_errors_total", "")
	assert.Equal(t, uint64(0), mc.Value)
	mc.Inc()
	assert.Equal(t, uint64(1), mc.Value)
	mc.Dec()
	assert.Equal(t, uint64(1), mc.Value)
	mc.Set(5)
	assert.Equal(t, uint64(1), mc.Get())

}

func TestStart(t *testing.T) {
	cfg := config.NewMockConfig()
	cfg.MGlobal = &config.Global{
		Status: config.Status{
			Addr: "127.0.0.1:8081",
		},
	}

	s := New(cfg)
	s.Start()
	time.Sleep(time.Millisecond * 500)

	resp, err := http.Get("http://localhost:8081/healthcheck")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "panoptes alive and reachable", string(body))

	resp, err = http.Get("http://localhost:8081/metrics")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}
