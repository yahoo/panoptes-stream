package status

import (
	"fmt"
	"net/http"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type Status struct {
	cfg             config.Config
	logger          *zap.Logger
	telemetryStatus *telemetry.Status
}

type healthcheck struct{}

func (h *healthcheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "panoptes alive and reachable")
}

func New(cfg config.Config, t *telemetry.Telemetry) *Status {
	return &Status{
		cfg:             cfg,
		logger:          cfg.Logger(),
		telemetryStatus: t.GetStatus(),
	}
}

func (t *Status) prometheus() {
	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "panoptes_connected_devices",
		Help: "",
	},
		func() float64 {
			return float64(t.telemetryStatus.ConnectedDevices)
		})

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "panoptes_total_devices",
		Help: "",
	},
		func() float64 {
			return float64(t.telemetryStatus.TotalDevices)
		})

	promauto.NewCounterFunc(prometheus.CounterOpts{
		Name: "panoptes_total_reconnect",
		Help: "",
	},
		func() float64 {
			return float64(t.telemetryStatus.Reconnect)
		})
}

func (s *Status) Start() {
	go func() {
		addr := s.cfg.Global().Status.Addr
		s.logger.Info("starting status server", zap.String("address", addr))

		s.prometheus()
		http.Handle("/metrics", promhttp.Handler())
		http.Handle("/healthcheck", new(healthcheck))
		http.ListenAndServe(addr, nil)
	}()
}
