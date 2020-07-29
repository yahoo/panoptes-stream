package gnmi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/ygot/ygot"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"git.vzbuilders.com/marshadrad/panoptes/config"
	"git.vzbuilders.com/marshadrad/panoptes/status"
	"git.vzbuilders.com/marshadrad/panoptes/telemetry"
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

	pathOutput map[string]string
}

// New creates a GNMI.
func New(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	var (
		subscriptions = []*gpb.Subscription{}
		pathOutput    = make(map[string]string)
		metrics       = make(map[string]status.Metrics)
	)

	metrics["gRPCDataTotal"] = status.NewCounter("generic_gnmi_grpc_data_total", "")
	metrics["dropsTotal"] = status.NewCounter("generic_gnmi_drops_total", "")
	metrics["errorsTotal"] = status.NewCounter("generic_gnmi_errors_total", "")
	metrics["processNSecond"] = status.NewGauge("generic_gnmi_process_nanosecond", "")

	status.Register(status.Labels{"host": conn.Target()}, metrics)

	for _, sensor := range sensors {
		path, _ := ygot.StringToPath(sensor.Path, ygot.StructuredPath, ygot.StringSlicePath)

		mode := gpb.SubscriptionMode_value[strings.ToUpper(sensor.Mode)]
		sampleInterval := time.Duration(sensor.SampleInterval) * time.Second
		heartbeatInterval := time.Duration(sensor.HeartbeatInterval) * time.Second
		subscriptions = append(subscriptions, &gpb.Subscription{
			Path:              path,
			Mode:              gpb.SubscriptionMode(mode),
			SampleInterval:    uint64(sampleInterval.Nanoseconds()),
			HeartbeatInterval: uint64(heartbeatInterval.Nanoseconds()),
			SuppressRedundant: sensor.SuppressRedundant,
		})

		pathOutput[sanitizePath(sensor.Path)] = sensor.Output
	}

	return &GNMI{
		logger:        logger,
		conn:          conn,
		subscriptions: subscriptions,
		dataChan:      make(chan *gpb.SubscribeResponse, 100),
		outChan:       outChan,
		pathOutput:    pathOutput,
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

	for i := 0; i < 1; i++ {
		go g.worker(ctx)
	}

	for ctx.Err() == nil {
		resp, err := subClient.Recv()
		if err != nil && ctx.Err() == nil {
			return err
		}

		if resp != nil {
			g.dataChan <- resp
			g.metrics["gRPCDataTotal"].Inc()
		}
	}

	return nil
}
func (g *GNMI) worker(ctx context.Context) {
	var start time.Time

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
				err := g.datastore(resp.Update.Prefix, update, resp.Update.Timestamp)
				if err != nil {
					g.logger.Error("generic.gnmi", zap.Error(err))
				}
			}

			g.metrics["processNSecond"].Set(uint64(time.Since(start).Nanoseconds()))

		case <-ctx.Done():
			return
		}
	}
}

func (g *GNMI) datastore(uPrefix *gpb.Path, update *gpb.Update, timestamp int64) error {
	var (
		path   []*gpb.PathElem
		labels map[string]string
		output string
		prefix string
		key    string
	)

	if uPrefix != nil && len(uPrefix.Elem) > 0 {
		path = append(path, uPrefix.Elem...)
		path = append(path, update.Path.Elem...)
	} else {
		path = update.Path.Elem
	}

	for i := 0; i < 2; i++ {
		labels = make(map[string]string)
		buf := bytes.NewBufferString("")

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
					prefix = buf.String() + "/"
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

	if output == "" {
		return errors.New("output not found")
	}

	value, err := getValue(update.Val)
	if err != nil {
		return err
	}

	systemID, _, _ := net.SplitHostPort(g.conn.Target())

	ds := telemetry.DataStore{
		"prefix":    prefix,
		"labels":    labels,
		"timestamp": timestamp,
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

func sanitizePath(path string) string {
	if !strings.HasSuffix(path, "/") {
		path = fmt.Sprintf("%s/", path)
	}

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	return path
}

func isRawRequested(output string) bool {
	return strings.HasSuffix(output, "::raw")
}

func Version() string {
	return gnmiVersion
}
