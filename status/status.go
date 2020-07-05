package status

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

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
		addr := s.cfg.Global().Status.Addr
		s.logger.Info("starting status server", zap.String("address", addr))

		http.Handle("/metrics", promhttp.Handler())
		http.Handle("/healthcheck", new(healthcheck))
		http.ListenAndServe(addr, nil)
	}()
}

func Register(metrics ...interface{}) {
	prefix := "panoptes_"

	for _, metric := range metrics {
		switch v := metric.(type) {
		case *MetricCounter:
			promauto.NewCounterFunc(prometheus.CounterOpts{
				Name: prefix + v.Name,
				Help: v.Help,
			},
				func() float64 {
					return float64(v.Value)
				})
		case *MetricGauge:
			promauto.NewGaugeFunc(prometheus.GaugeOpts{
				Name: prefix + v.Name,
				Help: v.Help,
			},
				func() float64 {
					return float64(v.Value)
				})
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

func (m *MetricGauge) Inc() {
	atomic.AddUint64(&m.Value, 1)
}

func (m *MetricGauge) Dec() {
	atomic.AddUint64(&m.Value, ^uint64(0))
}
