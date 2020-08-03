package status

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/secret"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Metrics interface {
	Dec()
	Inc()
	Set(uint64)
}

type Status struct {
	cfg    config.Config
	logger *zap.Logger
}

type Metric struct {
	Name string
	Help string
}

type MetricCounter struct {
	Metric
	Value uint64
}

type MetricGauge struct {
	Metric
	Value uint64
}

type Labels = prometheus.Labels

type healthcheck struct{}

func (h *healthcheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "panoptes alive and reachable")
}

func New(cfg config.Config) *Status {
	return &Status{
		cfg:    cfg,
		logger: cfg.Logger(),
	}
}

func (s *Status) Start() {
	go func() {
		if err := s.start(); err != nil {
			s.logger.Error("status", zap.Error(err))
		}
	}()
}

func (s *Status) start() error {
	config := s.cfg.Global().Status
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

func NewCounter(name, help string) *MetricCounter {
	return &MetricCounter{
		Metric: Metric{
			Name: name,
			Help: help,
		},
	}
}

func NewGauge(name, help string) *MetricGauge {
	return &MetricGauge{
		Metric: Metric{
			Name: name,
			Help: help,
		},
	}
}

func (m *MetricCounter) Inc() {
	atomic.AddUint64(&m.Value, 1)
}

func (m *MetricCounter) Dec() {
	// doesn't support
}

func (m *MetricCounter) Set(i uint64) {
	// doesn't support
}

func (m *MetricGauge) Inc() {
	atomic.AddUint64(&m.Value, 1)
}

func (m *MetricGauge) Dec() {
	atomic.AddUint64(&m.Value, ^uint64(0))
}

func (m *MetricGauge) Set(i uint64) {
	atomic.StoreUint64(&m.Value, i)
}
