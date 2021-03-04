//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package gnmi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/yahoo/panoptes-stream/config"
	"github.com/yahoo/panoptes-stream/status"
	"github.com/yahoo/panoptes-stream/telemetry"
	"github.com/yahoo/panoptes-stream/telemetry/juniper/proto/GnmiJuniperTelemetryHeader"
)

var gnmiVersion = "0.7.0"

// GNMI represents a GNMI Juniper.
type GNMI struct {
	conn          *grpc.ClientConn
	subscriptions []*gpb.Subscription

	dataChan chan *gpb.SubscribeResponse
	outChan  telemetry.ExtDSChan
	logger   *zap.Logger

	metrics map[string]status.Metrics

	pathOutput map[string]string
}

// New creates a GNMI.
func New(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	var metrics = make(map[string]status.Metrics)

	metrics["gRPCDataTotal"] = status.NewCounter("juniper_gnmi_grpc_data_total", "")
	metrics["dropsTotal"] = status.NewCounter("juniper_gnmi_drops_total", "")
	metrics["errorsTotal"] = status.NewCounter("juniper_gnmi_errors_total", "")
	metrics["processNSecond"] = status.NewGauge("juniper_gnmi_process_nanosecond", "")

	status.Register(status.Labels{"host": conn.Target()}, metrics)

	return &GNMI{
		logger:        logger,
		conn:          conn,
		subscriptions: telemetry.GetGNMISubscriptions(sensors),
		pathOutput:    telemetry.GetPathOutput(sensors),
		dataChan:      make(chan *gpb.SubscribeResponse, 100),
		outChan:       outChan,
		metrics:       metrics,
	}
}

// Start starts to get stream and fan-out to workers
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

	workers := config.GetEnvInt("JUNIPER_GNMI_WORKERS", 1)
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

			if err := g.datastore(buf, resp, systemID); err != nil {
				g.logger.Error("juniper.gnmi", zap.Error(err))
			}

			buf.Reset()

			g.metrics["processNSecond"].Set(uint64(time.Since(start).Nanoseconds()))

		case <-ctx.Done():
			return
		}
	}
}

func (g *GNMI) datastore(buf *bytes.Buffer, resp *gpb.SubscribeResponse_Update, systemID string) error {
	var (
		path, output string
		timestamp    interface{}
		labels       map[string]string
		ok           bool
	)

	prefix, prefixLabels := getPrefix(buf, resp.Update.Prefix.Elem)

	for _, update := range resp.Update.Update {
		buf.Reset()

		key, keyLabels := getKey(buf, update.Path.Elem)
		value, err := getValue(update.Val)
		if err != nil {
			continue
		}

		if key == "__juniper_telemetry_header__" {
			path, err = getPath(value)
			if err != nil {
				return err
			}
			continue
		}

		if key == "__timestamp__" {
			timestamp = resp.Update.GetTimestamp()
			continue
		}

		if strings.HasPrefix(key, "__") {
			continue
		}

		if output == "" && path != "" {
			output, ok = g.pathOutput[path]
			if !ok {
				return fmt.Errorf("out not found - %s", path)
			}
		}

		labels = telemetry.MergeLabels(keyLabels, prefixLabels)

		dataStore := telemetry.DataStore{
			"prefix":    prefix,
			"labels":    labels,
			"timestamp": timestamp,
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
			g.logger.Warn("juniper.gnmi", zap.String("error", "dataset drop"))
		}

	}

	return nil
}

func getPrefix(buf *bytes.Buffer, path []*gpb.PathElem) (string, map[string]string) {
	labels := make(map[string]string)

	for _, elem := range path {
		if len(elem.Name) > 0 {
			buf.WriteRune('/')
			buf.WriteString(elem.Name)
		}

		for key, value := range elem.Key {
			labels[key] = value
		}
	}

	return buf.String(), labels
}

func getKey(buf *bytes.Buffer, path []*gpb.PathElem) (string, map[string]string) {
	labels := make(map[string]string)

	for _, elem := range path {
		if len(elem.Name) > 0 {
			buf.WriteRune('/')
			buf.WriteString(elem.Name)
		}

		for key, value := range elem.Key {
			labels[key] = value
		}
	}

	buf.ReadRune()

	return buf.String(), labels
}

func getPath(value interface{}) (string, error) {
	var path string

	h, ok := value.(GnmiJuniperTelemetryHeader.GnmiJuniperTelemetryHeader)
	if ok {
		p := strings.Split(h.Path, ":")
		if len(p) > 1 {
			path = p[1]
		}
	}

	if path == "" {
		return path, fmt.Errorf("invalid juniper telemetry header")
	}

	return path, nil
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
	case *gpb.TypedValue_AnyVal:
		value, err = getAnyVal(tv.GetAnyVal())
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

func getAnyVal(anyMsg *anypb.Any) (interface{}, error) {
	anyMsgName, err := ptypes.AnyMessageName(anyMsg)
	if err != nil {
		return nil, fmt.Errorf("proto any message: %v", err)
	}
	if anyMsgName == "GnmiJuniperTelemetryHeader" {
		hdr := GnmiJuniperTelemetryHeader.GnmiJuniperTelemetryHeader{}
		if err := ptypes.UnmarshalAny(anyMsg, &hdr); err != nil {
			return nil, err
		}
		return hdr, nil
	}

	return nil, fmt.Errorf("proto any unknown msg type")
}

// Version returns version
func Version() string {
	return gnmiVersion
}
