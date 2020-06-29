package gnmi

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry/juniper/proto/GnmiJuniperTelemetryHeader"
	"go.uber.org/zap"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry"

	"github.com/golang/protobuf/ptypes"
	"github.com/openconfig/gnmi/path"
	"github.com/openconfig/ygot/ygot"
	"google.golang.org/grpc"

	apb "github.com/golang/protobuf/ptypes/any"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	_ "github.com/openconfig/gnmi/proto/gnmi_ext"

	"git.vzbuilders.com/marshadrad/panoptes/config"
)

var gnmiVersion = "0.7.0"

// GNMI ...
type GNMI struct {
	conn          *grpc.ClientConn
	subscriptions []*gpb.Subscription

	dataChan chan *gpb.SubscribeResponse
	outChan  telemetry.ExtDSChan
	lg       *zap.Logger

	pathOutput map[string]string
}

// New ...
func New(lg *zap.Logger, conn *grpc.ClientConn, sensors []*config.Sensor, outChan telemetry.ExtDSChan) telemetry.NMI {
	subscriptions := []*gpb.Subscription{}
	pathOutput := make(map[string]string)

	for _, sensor := range sensors {
		path, _ := ygot.StringToPath(sensor.Path, ygot.StructuredPath, ygot.StringSlicePath)

		mode := gpb.SubscriptionMode_value[strings.ToUpper(sensor.Mode)]
		sampleInterval := time.Duration(sensor.Interval) * time.Second
		subscriptions = append(subscriptions, &gpb.Subscription{
			Path:              path,
			Mode:              gpb.SubscriptionMode(mode),
			SampleInterval:    uint64(sampleInterval.Nanoseconds()),
			SuppressRedundant: false,
		})

		pathOutput[sensor.Path] = sensor.Output
	}

	return &GNMI{
		conn:          conn,
		subscriptions: subscriptions,
		dataChan:      make(chan *gpb.SubscribeResponse, 100),
		outChan:       outChan,
		pathOutput:    pathOutput,
	}
}

// Start ...
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
		if err != nil {
			return err
		}

		g.dataChan <- resp
	}

	return nil
}
func (g *GNMI) worker(ctx context.Context) {
	regxPath := regexp.MustCompile("/:(/.*/):")
	for {
		select {
		case d, ok := <-g.dataChan:
			if !ok {
				return
			}

			//TODO ext := d.GetExtension()

			switch resp := d.Response.(type) {
			case *gpb.SubscribeResponse_Update:
				ds := g.decoder(resp)
				jHPath := ds["__path__"].(string)
				delete(ds, "__path__")
				path := regxPath.FindStringSubmatch(jHPath)
				if len(path) > 1 {
					output, ok := g.pathOutput[path[1]]
					if !ok {
						g.lg.Warn("path to output not found", zap.String("path", jHPath))
						continue
					}

					select {
					case g.outChan <- telemetry.ExtDataStore{
						DS:     ds,
						Output: output,
					}:
					default:
					}

				} else {
					g.lg.Warn("path not found", zap.String("path", jHPath))
				}

			case *gpb.SubscribeResponse_SyncResponse:
				// TODO
			case *gpb.SubscribeResponse_Error:
				err := fmt.Errorf("%s", resp)
				g.lg.Error("error in sub response", zap.Error(err))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (g *GNMI) decoder(resp *gpb.SubscribeResponse_Update) telemetry.DataStore {
	ds := make(telemetry.DataStore)
	ds["__service__"] = fmt.Sprintf("gnmi_v%s", gnmiVersion)

	ds["__update_timestamp__"] = resp.Update.GetTimestamp()
	ds["__prefix__"] = path.ToStrings(resp.Update.GetPrefix(), true)

	for _, update := range resp.Update.Update {
		var value interface{}
		var jsondata []byte

		pathSlice := path.ToStrings(update.Path, false)
		key := strings.Join(pathSlice, "/")

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
		case *gpb.TypedValue_AnyVal:
			value = val.AnyVal
			anyMsg := value.(*apb.Any)
			anyMsgName, err := ptypes.AnyMessageName(anyMsg)
			if err != nil {
				g.lg.Error("proto any message invalid", zap.Error(err))
			}
			if anyMsgName == "GnmiJuniperTelemetryHeader" {
				hdr := GnmiJuniperTelemetryHeader.GnmiJuniperTelemetryHeader{}
				ptypes.UnmarshalAny(anyMsg, &hdr)
				value = hdr
				ds["__path__"] = hdr.Path
			}
		case *gpb.TypedValue_LeaflistVal:
			// TODO
		case *gpb.TypedValue_JsonIetfVal:
			jsondata = val.JsonIetfVal
		case *gpb.TypedValue_JsonVal:
			jsondata = val.JsonVal
		}

		if value != nil {
			ds[key] = value
		} else if jsondata != nil {
			// TODO
		}

	}

	//TODO ADD OUTPUT Info

	return ds
}

func Version() string {
	return gnmiVersion
}