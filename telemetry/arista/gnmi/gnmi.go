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

// GNMI represents a GNMI Juniper.
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

		if strings.HasSuffix(sensor.Path, "/") {
			pathOutput[sensor.Path] = sensor.Output
		} else {
			pathOutput[fmt.Sprintf("%s/", sensor.Path)] = sensor.Output
		}
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
	systemID, _, _ := net.SplitHostPort(g.conn.Target())

	for {
		select {
		case d, ok := <-g.dataChan:
			if !ok {
				return
			}

			switch resp := d.Response.(type) {
			case *gpb.SubscribeResponse_Update:
				ds, err := g.dataStore(resp)
				if err != nil {
					continue
				}

				ds["__timestamp__"] = resp.Update.Timestamp
				ds["__system_id__"] = systemID
				ds["__service__"] = "arista.gnmi"

				output, err := g.findOutput(resp)
				if err != nil {
					g.logger.Error("arista.gnmi", zap.Error(err))
				}

				select {
				case g.outChan <- telemetry.ExtDataStore{
					DS:     ds,
					Output: output,
				}:
				default:
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

func (g *GNMI) dataStore(resp *gpb.SubscribeResponse_Update) (telemetry.DataStore, error) {
	ds := make(telemetry.DataStore)

	for _, update := range resp.Update.Update {
		var value interface{}
		var jsondata []byte

		key, err := ygot.PathToString(update.Path)
		if err != nil {
			return nil, err
		}

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
			ds[key] = value
		} else if jsondata != nil {
			// TODO
			g.logger.Warn("JSON")
		}

	}

	return ds, nil
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
