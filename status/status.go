//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package status

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
)

// Metrics represents counter and gauge metrics.
type Metrics interface {
	Dec()
	Inc()
	Get() uint64
	Set(uint64)
}

// Status represents Panoptes status and healthcheck
type Status struct {
	cfg    config.Config
	logger *zap.Logger
}

// Metric represents a metric
type Metric struct {
	Name string
	Help string
}

// MetricCounter represents counter metric
type MetricCounter struct {
	Metric
	Value uint64
}

// MetricGauge represents gauge metric
type MetricGauge struct {
	Metric
	Value uint64
}

// Labels represents prometheus labels
type Labels = prometheus.Labels

type healthcheck struct{}

func (h *healthcheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "panoptes alive and reachable")
}

// New constructs a new status
func New(cfg config.Config) *Status {
	return &Status{
		cfg:    cfg,
		logger: cfg.Logger(),
	}
}

// Start starts status web service and exposes
// panoptes metrics and healthcheck
func (s *Status) Start() {
	go func() {
		if err := s.start(); err != nil {
			s.logger.Error("status", zap.Error(err))
		}
	}()
}

func (s *Status) start() error {
	config := s.cfg.Global().Status

	if len(config.Addr) < 1 {
		config.Addr = ":8081"
	}

	s.logger.Info("status", zap.String("address", config.Addr), zap.Bool("tls", config.TLSConfig.Enabled))

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/healthcheck", new(healthcheck))

	if !config.TLSConfig.Enabled {
		return http.ListenAndServe(config.Addr, nil)
	}

	tlsConfig, err := secret.GetTLSConfig(&config.TLSConfig)
	if err != nil {
		return err
	}

	srv := http.Server{
		Addr:      config.Addr,
		TLSConfig: tlsConfig,
	}

	return srv.ListenAndServeTLS("", "")
}

// Register registers a metric to prometheus
func Register(labels Labels, metrics map[string]Metrics) {
	prefix := "panoptes_"

	for _, metric := range metrics {
		switch v := metric.(type) {
		case *MetricCounter:
			prometheus.Register(prometheus.NewCounterFunc(prometheus.CounterOpts{
				Name:        prefix + v.Name,
				Help:        v.Help,
				ConstLabels: labels,
			},
				func() float64 {
					return float64(atomic.LoadUint64(&v.Value))
				}))
		case *MetricGauge:
			prometheus.Register(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
				Name:        prefix + v.Name,
				Help:        v.Help,
				ConstLabels: labels,
			},
				func() float64 {
					return float64(atomic.LoadUint64(&v.Value))
				}))
		}
	}
}

// Unregister unregisters a metric from prometheus
func Unregister(labels Labels, metrics map[string]Metrics) {
	prefix := "panoptes_"

	for _, metric := range metrics {
		switch v := metric.(type) {
		case *MetricCounter:
			prometheus.Unregister(prometheus.NewCounterFunc(prometheus.CounterOpts{
				Name:        prefix + v.Name,
				ConstLabels: labels,
			},
				func() float64 {
					return float64(v.Value)
				}))
		case *MetricGauge:
			prometheus.Unregister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
				Name:        prefix + v.Name,
				ConstLabels: labels,
			},
				func() float64 {
					return float64(v.Value)
				}))
		}
	}

}

// NewCounter creates a counter metric
func NewCounter(name, help string) *MetricCounter {
	return &MetricCounter{
		Metric: Metric{
			Name: name,
			Help: help,
		},
	}
}

// NewGauge creates a gauge metric
func NewGauge(name, help string) *MetricGauge {
	return &MetricGauge{
		Metric: Metric{
			Name: name,
			Help: help,
		},
	}
}

// Inc increases one unit counter metric
func (m *MetricCounter) Inc() {
	atomic.AddUint64(&m.Value, 1)
}

// Dec is not available for counter metric
func (m *MetricCounter) Dec() {
	// doesn't support
}

// Set is not available for counter metric
func (m *MetricCounter) Set(i uint64) {
	// doesn't support
}

// Get returns counter metric value
func (m *MetricCounter) Get() uint64 {
	return atomic.LoadUint64(&m.Value)
}

// Inc increases one unit to gauge metric
func (m *MetricGauge) Inc() {
	atomic.AddUint64(&m.Value, 1)
}

// Dec decreases one unit from gauge metric
func (m *MetricGauge) Dec() {
	atomic.AddUint64(&m.Value, ^uint64(0))
}

// Set sets gauge metric value
func (m *MetricGauge) Set(i uint64) {
	atomic.StoreUint64(&m.Value, i)
}

// Get returns gauge metric value
func (m *MetricGauge) Get() uint64 {
	return atomic.LoadUint64(&m.Value)
}
