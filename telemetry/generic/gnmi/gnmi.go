package gnmi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

		pathOutput[pathToString(path)] = sensor.Output
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
	for {
		select {
		case d, ok := <-g.dataChan:
			if !ok {
				return
			}

			resp, ok := d.Response.(*gpb.SubscribeResponse_Update)
			if !ok {
				continue
			}

			if len(resp.Update.Update) == 1 {
				err := g.signleUpdate(resp)
				if err != nil {
					g.logger.Error("generic.gnmi", zap.Error(err))
				}
			} else {
				//g.multiUpdates(resp)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (g *GNMI) signleUpdate(resp *gpb.SubscribeResponse_Update) error {
	var (
		labels map[string]string
		path   []*gpb.PathElem
		output string
		prefix string
		key    string
	)

	if resp.Update.Prefix != nil && len(resp.Update.Prefix.Elem) > 0 {
		path = append(resp.Update.Prefix.Elem, resp.Update.Update[0].Path.Elem...)
	} else {
		path = resp.Update.Update[0].Path.Elem
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
				log.Println("PREFIX", buf.String()+"/")
				p, ok := g.pathOutput[buf.String()+"/"]
				if ok {
					output = p
					prefix = buf.String() + "/"
					buf.Reset()
				}
			}

		}

		key = buf.String()

		if output != "" {
			break
		}

	}

	if output == "" {
		return errors.New("output not found")
	}

	value, err := getValue(resp.Update.Update[0].Val)
	if err != nil {
		return err
	}

	systemID, _, _ := net.SplitHostPort(g.conn.Target())

	ds := telemetry.DataStore{
		"__prefix":    prefix,
		"__labels":    labels,
		"__timestamp": resp.Update.Timestamp,
		"__system_id": systemID,

		key: value,
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

// pathToString converts path to string w/o keys and values
func pathToString(path *gpb.Path) string {
	buf := bytes.NewBufferString("")
	for _, elem := range path.Elem {
		if len(elem.Name) > 0 {
			buf.WriteRune('/')
			buf.WriteString(elem.Name)
		}

		for key, value := range elem.Key {
			buf.WriteRune('[')
			buf.WriteString(key + "=" + value)
			buf.WriteRune(']')
		}
	}

	buf.WriteRune('/')

	return buf.String()
}

func isRawRequested(output string) bool {
	return strings.HasSuffix(output, "::raw")
}

func Version() string {
	return gnmiVersion
}
