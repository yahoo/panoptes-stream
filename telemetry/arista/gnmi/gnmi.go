package gnmi

import (
	"bytes"
	"context"
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

	pathOutput map[string]string
}

// New creates a GNMI.
func New(logger *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	subscriptions := []*gpb.Subscription{}
	pathOutput := make(map[string]string)

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
			SuppressRedundant: false,
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
					g.logger.Error("arista.gnmi", zap.Error(err))
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

		value := getValue(update)
		key = strings.Replace(key, prefix, "", -1)

		ds := telemetry.DataStore{
			"__prefix__":    prefix,
			"__labels__":    labels,
			"__timestamp__": resp.Update.Timestamp,
			"__system_id__": systemID,

			key: value,
		}

		select {
		case g.outChan <- telemetry.ExtDataStore{
			DS:     ds,
			Output: output,
		}:
		default:
			g.logger.Warn("drop!")
		}
	}
}

func (g *GNMI) rawDataStore(resp *gpb.SubscribeResponse_Update, output string) {
	ds := make(telemetry.DataStore)

	for _, update := range resp.Update.Update {
		key, err := ygot.PathToString(update.Path)
		if err != nil {
			g.logger.Error("arista.gnmi", zap.Error(err))
			return
		}

		value := getValue(update)
		ds[key] = value
	}

	systemID, _, _ := net.SplitHostPort(g.conn.Target())

	ds["__timestamp__"] = resp.Update.Timestamp
	ds["__system_id__"] = systemID
	ds["__service__"] = "arista.gnmi"

	select {
	case g.outChan <- telemetry.ExtDataStore{
		DS:     ds,
		Output: output,
	}:
	default:
		g.logger.Warn("drop!")
	}
}

func getValue(update *gpb.Update) interface{} {
	var jsondata []byte
	var value interface{}

	switch val := update.Val.Value.(type) {
	case *gpb.TypedValue_AsciiVal:
		value = val.AsciiVal
	case *gpb.TypedValue_BoolVal:
		value = val.BoolVal
	case *gpb.TypedValue_BytesVal:
		value = val.BytesVal
	case *gpb.TypedValue_DecimalVal:
		value = float64(val.DecimalVal.Digits) / math.Pow(10, float64(val.DecimalVal.Precision))
	case *gpb.TypedValue_FloatVal:
		value = val.FloatVal
	case *gpb.TypedValue_IntVal:
		value = val.IntVal
	case *gpb.TypedValue_StringVal:
		value = val.StringVal
	case *gpb.TypedValue_UintVal:
		value = val.UintVal
	case *gpb.TypedValue_JsonIetfVal:
		jsondata = val.JsonIetfVal
	case *gpb.TypedValue_JsonVal:
		jsondata = val.JsonVal
	}

	if value != nil {
		return value
	} else if jsondata != nil {
		// TODO
		panic("JSON")
	}

	return nil
}

func Version() string {
	return gnmiVersion
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
