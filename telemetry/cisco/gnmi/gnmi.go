//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package gnmi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/status"
	"github.com/yahoo/panoptes-stream/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var gnmiVersion = "0.0.1"

// GNMI represents a GNMI.
type GNMI struct {
	conn          *grpc.ClientConn
	subscriptions []*gpb.Subscription

	dataChan chan *gpb.SubscribeResponse
	outChan  telemetry.ExtDSChan
	logger   *zap.Logger

	metrics map[string]status.Metrics

	pathOutput    map[string]string
	defaultOutput string
}

// New creates a GNMI.
func New(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	var metrics = make(map[string]status.Metrics)

	metrics["gRPCDataTotal"] = status.NewCounter("cisco_gnmi_grpc_data_total", "")
	metrics["dropsTotal"] = status.NewCounter("cisco_gnmi_drops_total", "")
	metrics["errorsTotal"] = status.NewCounter("cisco_gnmi_errors_total", "")
	metrics["processNSecond"] = status.NewGauge("cisco_gnmi_process_nanosecond", "")

	status.Register(status.Labels{"host": conn.Target()}, metrics)

	return &GNMI{
		logger:        logger,
		conn:          conn,
		subscriptions: telemetry.GetGNMISubscriptions(sensors),
		dataChan:      make(chan *gpb.SubscribeResponse, 100),
		outChan:       outChan,
		pathOutput:    telemetry.GetPathOutput(sensors),
		defaultOutput: telemetry.GetDefaultOutput(sensors),
		metrics:       metrics,
	}
}

// Start gets stream metrics and fan-out to workers
func (g *GNMI) Start(ctx context.Context) error {
	defer status.Unregister(status.Labels{"host": g.conn.Target()}, g.metrics)

	client := gpb.NewGNMIClient(g.conn)
	subReq := &gpb.SubscribeRequest{
		Request: &gpb.SubscribeRequest_Subscribe{
			Subscribe: &gpb.SubscriptionList{
				Mode:         gpb.SubscriptionList_STREAM,
				Encoding:     gpb.Encoding(gpb.Encoding_value["PROTO"]),
				Subscription: g.subscriptions,
				UpdatesOnly:  false,
			},
		},
	}

	subClient, err := client.Subscribe(ctx)
	if err != nil {
		return err
	}

	err = subClient.Send(subReq)
	if err != nil {
		return err
	}

	workers := config.GetEnvInt("CISCO_GNMI_WORKERS", 1)
	for i := 0; i < workers; i++ {
		go g.worker(ctx)
	}

	for {
		resp, err := subClient.Recv()
		if err != nil && ctx.Err() == nil {
			return err
		}

		if ctx.Err() != nil {
			return nil
		}

		g.dataChan <- resp
		g.metrics["gRPCDataTotal"].Inc()
	}

}

func (g *GNMI) worker(ctx context.Context) {
	var (
		start          time.Time
		buf            = new(bytes.Buffer)
		systemID, _, _ = net.SplitHostPort(g.conn.Target())
	)

	for {
		select {
		case d, ok := <-g.dataChan:
			if !ok {
				return
			}

			start = time.Now()

			resp, ok := d.Response.(*gpb.SubscribeResponse_Update)
			if !ok {
				continue
			}

			if err := respValidation(resp); err != nil {
				g.logger.Error("cisco.gnmi", zap.Error(err))
				continue
			}

			if err := g.datastore(buf, resp.Update, systemID); err != nil {
				g.logger.Error("cisco.gnmi", zap.Error(err))
			}

			buf.Reset()

			g.metrics["processNSecond"].Set(uint64(time.Since(start).Nanoseconds()))

		case <-ctx.Done():
			return
		}
	}
}

func (g *GNMI) datastore(buf *bytes.Buffer, n *gpb.Notification, systemID string) error {
	var labels map[string]string

	prefix, prefixLabels, output := g.getPrefix(buf, n.Prefix)

	if g.defaultOutput != "" {
		output = g.defaultOutput
	} else if output == "" {
		return errors.New("output not found")
	}

	for _, update := range n.Update {
		buf.Reset()

		key, keyLabels := telemetry.GetKey(buf, update.Path.Elem)
		value, err := telemetry.GetValue(update.Val)
		if err != nil {
			continue
		}

		labels = telemetry.MergeLabels(keyLabels, prefixLabels, prefix)

		dataStore := telemetry.DataStore{
			"prefix":    prefix,
			"labels":    labels,
			"timestamp": n.Timestamp,
			"system_id": systemID,
			"key":       key,
			"value":     value,
		}

		select {
		case g.outChan <- telemetry.ExtDataStore{
			DS:     dataStore,
			Output: output,
		}:
		default:
			g.metrics["dropsTotal"].Inc()
			g.logger.Warn("cisco.gnmi", zap.String("error", "dataset drop"))
		}
	}

	return nil
}

func (g *GNMI) getPrefix(buf *bytes.Buffer, path *gpb.Path) (string, map[string]string, string) {
	labels := make(map[string]string)
	var output, prefix string

	for i := 0; i < 2; i++ {
		for _, elem := range path.Elem {
			if len(elem.Name) > 0 {
				buf.WriteRune('/')
				buf.WriteString(elem.Name)
			}

			for key, value := range elem.Key {
				labels[key] = value
				if i == 1 {
					buf.WriteString(fmt.Sprintf("[%s=%s]", key, value))
				}
			}
		}

		if len(prefix) < 1 {
			prefix = buf.String()
		}

		if v, ok := g.pathOutput[buf.String()+"/"]; ok {
			output = v
			break
		}

		buf.Reset()
	}

	return prefix, labels, output
}

func respValidation(resp *gpb.SubscribeResponse_Update) error {
	if resp.Update.Prefix == nil {
		return errors.New("invalid cisco response")
	}

	return nil
}

// Version returns version
func Version() string {
	return gnmiVersion
}
