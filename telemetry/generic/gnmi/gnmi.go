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

var (
	gnmiVersion = "0.0.1"

	metricGRPCDataTotal  = status.NewCounter("generic_gnmi_grpc_data_total", "")
	metricGNMIDropsTotal = status.NewCounter("generic_gnmi_drops_total", "")
)

// GNMI represents a GNMI.
type GNMI struct {
	conn          *grpc.ClientConn
	subscriptions []*gpb.Subscription

	dataChan chan *gpb.SubscribeResponse
	outChan  telemetry.ExtDSChan
	logger   *zap.Logger

	pathOutput map[string]string
}

// New creates a GNMI.
func New(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	var (
		subscriptions = []*gpb.Subscription{}
		pathOutput    = make(map[string]string)
	)

	status.Register(
		status.Labels{"host": conn.Target()},
		metricGRPCDataTotal,
		metricGNMIDropsTotal,
	)

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
	}
}

// Start starts to get stream and fan-out to workers
func (g *GNMI) Start(ctx context.Context) error {
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
			metricGRPCDataTotal.Inc()
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

			switch resp := d.Response.(type) {
			case *gpb.SubscribeResponse_Update:
				output, err := g.findOutput(resp)
				if err != nil {
					g.logger.Error("generic.gnmi", zap.Error(err))
					continue
				}

				if isRawRequested(output) {
					g.rawDataStore(resp, output)
				} else {
					g.dataStore(resp, output)
				}

			case *gpb.SubscribeResponse_SyncResponse:
				// TODO
			case *gpb.SubscribeResponse_Error:
				err := fmt.Errorf("%s", resp)
				g.logger.Error("error in sub response", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (g *GNMI) dataStore(resp *gpb.SubscribeResponse_Update, output string) {
	systemID, _, _ := net.SplitHostPort(g.conn.Target())

	for _, update := range resp.Update.Update {
		labels, prefix, key := g.parsePath(update)

		value, err := getValue(update.Val)
		if err != nil {
			g.logger.Error("generic.gnmi", zap.Error(err))
			continue
		}
		key = strings.Replace(key, prefix, "", -1)

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
			metricGNMIDropsTotal.Inc()
			g.logger.Warn("generic.gnmi", zap.String("error", "dataset drop"))
		}
	}
}

func (g *GNMI) rawDataStore(resp *gpb.SubscribeResponse_Update, output string) {
	ds := make(telemetry.DataStore)

	for _, update := range resp.Update.Update {
		key, err := ygot.PathToString(update.Path)
		if err != nil {
			g.logger.Error("generic.gnmi", zap.Error(err))
			continue
		}

		value, err := getValue(update.Val)
		if err != nil {
			g.logger.Error("generic.gnmi", zap.Error(err))
			continue
		}
		ds[key] = value
	}

	systemID, _, _ := net.SplitHostPort(g.conn.Target())

	ds["__timestamp__"] = resp.Update.Timestamp
	ds["__system_id__"] = systemID
	ds["__service__"] = "generic.gnmi"

	select {
	case g.outChan <- telemetry.ExtDataStore{
		DS:     ds,
		Output: output,
	}:
	default:
		metricGNMIDropsTotal.Inc()
		g.logger.Warn("generic.gnmi", zap.String("error", "dataset drop"))
	}
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

func (g *GNMI) findOutput(resp *gpb.SubscribeResponse_Update) (string, error) {
	if len(resp.Update.Update) < 1 {
		return "", errors.New("update is empty")
	}

	path := resp.Update.Update[0].Path
	buf := bytes.NewBufferString("")

	for _, elem := range path.Elem {
		if len(elem.Name) > 0 {
			buf.WriteRune('/')
			buf.WriteString(elem.Name)
		}
		p, ok := g.pathOutput[buf.String()+"/"]
		if ok {
			return p, nil
		}
	}

	return "", errors.New("path to output not found")
}

func (g *GNMI) parsePath(update *gpb.Update) (map[string]string, string, string) {
	Labels := make(map[string]string)
	buf := bytes.NewBufferString("")
	prefix := ""

	for _, elem := range update.Path.Elem {
		if len(elem.Name) > 0 {
			buf.WriteRune('/')
			buf.WriteString(elem.Name)
		}

		_, ok := g.pathOutput[buf.String()+"/"]
		if ok {
			prefix = buf.String() + "/"
		}

		for key, value := range elem.Key {
			if _, ok := Labels[key]; ok {
				Labels[buf.String()+"/"+key] = value
			} else {
				Labels[key] = value
			}

		}
	}

	return Labels, prefix, buf.String()
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
