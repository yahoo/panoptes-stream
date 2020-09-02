//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package gnmi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"time"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
)

var gnmiVersion = "0.0.1"

// GNMI represents a gNMI for Arista EOS telemetry.
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

// New creates a gNMI and register proper metrics.
func New(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	var metrics = make(map[string]status.Metrics)

	metrics["gRPCDataTotal"] = status.NewCounter("arista_gnmi_grpc_data_total", "")
	metrics["dropsTotal"] = status.NewCounter("arista_gnmi_drops_total", "")
	metrics["errorsTotal"] = status.NewCounter("arista_gnmi_errors_total", "")
	metrics["processNSecond"] = status.NewGauge("arista_gnmi_process_nanosecond", "")

	status.Register(status.Labels{"host": conn.Target()}, metrics)

	return &GNMI{
		logger:        logger,
		conn:          conn,
		subscriptions: telemetry.GetGNMISubscriptions(sensors),
		pathOutput:    telemetry.GetPathOutput(sensors),
		defaultOutput: telemetry.GetDefaultOutput(sensors),
		dataChan:      make(chan *gpb.SubscribeResponse, 100),
		outChan:       outChan,
		metrics:       metrics,
	}
}

// Start starts to get stream and fan-out to workers.
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

	for i := 0; i < 1; i++ {
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

			for _, update := range resp.Update.Update {
				err := g.datastore(buf, resp.Update, update, systemID)
				if err != nil {
					g.logger.Error("arista.gnmi", zap.Error(err))
				}
			}

			g.metrics["processNSecond"].Set(uint64(time.Since(start).Nanoseconds()))

		case <-ctx.Done():
			return
		}
	}
}

func (g *GNMI) datastore(buf *bytes.Buffer, n *gpb.Notification, update *gpb.Update, systemID string) error {
	var (
		path   []*gpb.PathElem
		labels map[string]string
		output string
		prefix string
		key    string
	)

	if n.Prefix != nil && len(n.Prefix.Elem) > 0 {
		path = n.Prefix.Elem
	}

	path = append(path, update.Path.Elem...)

	for i := 0; i < 2; i++ {
		labels = make(map[string]string)
		buf.Reset()

		for _, elem := range path {
			if len(elem.Name) > 0 {
				buf.WriteRune('/')
				buf.WriteString(elem.Name)
			}

			// labels
			for key, value := range elem.Key {
				if _, ok := labels[key]; ok {
					labels[buf.String()+"/"+key] = value
				} else {
					labels[key] = value
					if i == 1 {
						buf.WriteString(fmt.Sprintf("[%s=%s]", key, value))
					}
				}
			}

			// prefix
			if output == "" {
				p, ok := g.pathOutput[buf.String()+"/"]
				if ok {
					output = p
					prefix = buf.String()
					buf.Reset()
				}
			}

		}

		buf.ReadRune()
		key = buf.String()

		if output != "" {
			break
		}
	}

	if g.defaultOutput != "" {
		output = g.defaultOutput
	} else if output == "" {
		return errors.New("output not found")
	}

	value, err := getValue(update.Val)
	if err != nil {
		return err
	}

	ds := telemetry.DataStore{
		"prefix":    prefix,
		"labels":    labels,
		"timestamp": n.Timestamp,
		"system_id": systemID,
		"key":       key,
		"value":     value,
	}

	select {
	case g.outChan <- telemetry.ExtDataStore{
		DS:     ds,
		Output: output,
	}:
	default:
		g.metrics["dropsTotal"].Inc()
		return errors.New("dataset drop")
	}

	return nil
}

func getValue(tv *gpb.TypedValue) (interface{}, error) {
	var (
		jsondata []byte
		value    interface{}
		err      error
	)

	switch tv.Value.(type) {
	case *gpb.TypedValue_AsciiVal:
		value = tv.GetAsciiVal()
	case *gpb.TypedValue_BoolVal:
		value = tv.GetBoolVal()
	case *gpb.TypedValue_BytesVal:
		value = tv.GetBytesVal()
	case *gpb.TypedValue_DecimalVal:
		value = float64(tv.GetDecimalVal().Digits) / math.Pow(10, float64(tv.GetDecimalVal().Precision))
	case *gpb.TypedValue_FloatVal:
		value = tv.GetFloatVal()
	case *gpb.TypedValue_IntVal:
		value = tv.GetIntVal()
	case *gpb.TypedValue_StringVal:
		value = tv.GetStringVal()
	case *gpb.TypedValue_UintVal:
		value = tv.GetUintVal()
	case *gpb.TypedValue_JsonIetfVal:
		jsondata = tv.GetJsonIetfVal()
	case *gpb.TypedValue_JsonVal:
		jsondata = tv.GetJsonVal()
	case *gpb.TypedValue_LeaflistVal:
		elems := tv.GetLeaflistVal().GetElement()
		value, err = getLeafList(elems)
	default:
		err = fmt.Errorf("unknown value type %+v", tv.Value)
	}

	if jsondata != nil {
		err = json.Unmarshal(jsondata, &value)
	}

	return value, err
}

func getLeafList(elems []*gpb.TypedValue) (interface{}, error) {
	list := []interface{}{}
	for _, v := range elems {
		ev, err := getValue(v)
		if err != nil {
			return nil, fmt.Errorf("leaflist error: %v", err)
		}
		list = append(list, ev)
	}

	return list, nil
}

// Version returns the current package version.
func Version() string {
	return gnmiVersion
}
