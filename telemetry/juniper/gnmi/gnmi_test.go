package gnmi

import (
	"context"
	"testing"
	"time"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry/generic/gnmi/mock"
)

func TestJuniperCountersMock(t *testing.T) {
	var (
		addr    = "127.0.0.1:50500"
		ch      = make(telemetry.ExtDSChan, 10)
		sensors []*config.Sensor
	)
	ln, err := mock.StartGNMIServer(addr, mock.Update{Notification: mock.JuniperUpdate(), Attempt: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	cfg := &config.MockConfig{}

	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*500)

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	sensors = append(sensors, &config.Sensor{
		Service: "juniper.gnmi",
		Output:  "console::stdout",
		Path:    "/interfaces/interface/state/counters",
	})

	g := New(cfg.Logger(), conn, sensors, ch)
	g.Start(ctx)

	tt := []struct {
		key   string
		value interface{}
	}{
		{
			key:   "state/counters/in-pkts",
			value: int64(23004050),
		},
		{
			key:   "state/counters/out-pkts",
			value: int64(23004056),
		},
		{
			key:   "state/counters/in-octets",
			value: int64(50302030597),
		},
		{
			key:   "state/counters/out-octets",
			value: int64(50302030492),
		},
		{
			key:   "state/counters/last-clear",
			value: "Never",
		},
	}

	r := make(map[string]telemetry.ExtDataStore)

	for i := 0; i < 5; i++ {
		select {
		case resp := <-ch:
			r[resp.DS["key"].(string)] = resp

			assert.Equal(t, resp.DS["prefix"].(string), "/interfaces/interface")
			assert.Equal(t, resp.DS["system_id"].(string), "127.0.0.1")
			assert.Equal(t, resp.DS["timestamp"].(int64), int64(1595951912880990837))
			assert.Equal(t, resp.DS["labels"].(map[string]string)["name"], "lo0")

		case <-ctx.Done():
			assert.Error(t, ctx.Err(), "timeout")
			return
		}
	}

	for _, test := range tt {
		resp := r[test.key]
		assert.Equal(t, resp.DS["value"], test.value)
	}
}

func TestGNMI_splitRawDataStore(t *testing.T) {
	type fields struct {
		conn          *grpc.ClientConn
		subscriptions []*gpb.Subscription
		dataChan      chan *gpb.SubscribeResponse
		outChan       telemetry.ExtDSChan
		logger        *zap.Logger
		metrics       map[string]status.Metrics
		pathOutput    map[string]string
	}
	type args struct {
		ds     telemetry.DataStore
		output string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GNMI{
				conn:          tt.fields.conn,
				subscriptions: tt.fields.subscriptions,
				dataChan:      tt.fields.dataChan,
				outChan:       tt.fields.outChan,
				logger:        tt.fields.logger,
				metrics:       tt.fields.metrics,
				pathOutput:    tt.fields.pathOutput,
			}
			g.splitRawDataStore(tt.args.ds, tt.args.output)
		})
	}
}
