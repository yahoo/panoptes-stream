package status

import (
	"fmt"
	"net/http"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Status struct {
	telemetryStatus *telemetry.Status
}

type healthcheck struct{}

func (h *healthcheck) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "panoptes alive and reachable")
}

func New(t *telemetry.Telemetry) *Status {
	return &Status{
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
		s.prometheus()
		http.Handle("/metrics", promhttp.Handler())
		http.Handle("/healthcheck", new(healthcheck))
		http.ListenAndServe("localhost:8081", nil)
	}()
}
