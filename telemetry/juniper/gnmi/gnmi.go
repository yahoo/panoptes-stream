package gnmi

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
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

var (
	gnmiVersion = "0.7.0"
	labelsRegex = regexp.MustCompile("(\\/[^\\/]*)\\[([A-Za-z0-9\\-\\/]*\\=[^\\[]*)\\]")
)

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
		subscriptions = append(subscriptions, &gpb.Subscription{
			Path:              path,
			Mode:              gpb.SubscriptionMode(mode),
			SampleInterval:    uint64(sampleInterval.Nanoseconds()),
			SuppressRedundant: sensor.SuppressRedundant,
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

	for i := 0; i < 4; i++ {
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

			//TODO ext := d.GetExtension()

			switch resp := d.Response.(type) {
			case *gpb.SubscribeResponse_Update:
				ds := g.rawDataStore(resp)

				path, err := getPath(ds)
				if err != nil {
					g.logger.Warn("path not found")
					continue
				}

				output, ok := g.pathOutput[path]
				if !ok {
					g.logger.Warn("path to output not found", zap.String("path", path))
					continue
				}

				if isRawRequested(output) {
					select {
					case g.outChan <- telemetry.ExtDataStore{
						DS:     ds,
						Output: output,
					}:
					default:
						g.logger.Warn("juniper.gnmi", zap.String("error", "dataset drop"))
					}
				} else {
					g.splitRawDataStore(ds, output)
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

func (g *GNMI) rawDataStore(resp *gpb.SubscribeResponse_Update) telemetry.DataStore {
	ds := make(telemetry.DataStore)
	ds["__service__"] = fmt.Sprintf("gnmi_v%s", gnmiVersion)

	ds["__update_timestamp__"] = resp.Update.GetTimestamp()
	ds["__prefix__"], _ = ygot.PathToString(resp.Update.GetPrefix())

	for _, update := range resp.Update.Update {

		pathSlice := path.ToStrings(update.Path, false)
		key := strings.Join(pathSlice, "/")

		value := g.getValue(update)
		ds[key] = value

	}

	return ds
}

func (g *GNMI) splitRawDataStore(ds telemetry.DataStore, output string) {
	labels, prefix := getLabels(ds["__prefix__"].(string))
	systemID, _, _ := net.SplitHostPort(g.conn.Target())

	for key, value := range ds {
		if !strings.HasPrefix(key, "__") {
			dataStore := telemetry.DataStore{
				"__prefix":    prefix,
				"__labels":    labels,
				"__timestamp": ds["__update_timestamp__"],
				"__system_id": systemID,

				key: value,
			}

			select {
			case g.outChan <- telemetry.ExtDataStore{
				DS:     dataStore,
				Output: output,
			}:
			default:
				g.logger.Warn("juniper.gnmi", zap.String("error", "dataset drop"))
			}
		}
	}

}

func getPath(ds telemetry.DataStore) (string, error) {
	regxPath := regexp.MustCompile("/:(/.*/):")

	if _, ok := ds["__juniper_telemetry_header__"]; ok {
		h := ds["__juniper_telemetry_header__"].(GnmiJuniperTelemetryHeader.GnmiJuniperTelemetryHeader)
		path := regxPath.FindStringSubmatch(h.Path)
		if len(path) > 1 {
			return path[1], nil
		}
	}

	return "", errors.New("path not found")
}

func (g *GNMI) getValue(update *gpb.Update) interface{} {
	var (
		jsondata []byte
		value    interface{}
	)

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
			g.logger.Error("proto any message invalid", zap.Error(err))
		}
		if anyMsgName == "GnmiJuniperTelemetryHeader" {
			hdr := GnmiJuniperTelemetryHeader.GnmiJuniperTelemetryHeader{}
			ptypes.UnmarshalAny(anyMsg, &hdr)
			value = hdr
		}
	case *gpb.TypedValue_LeaflistVal:
		// TODO
	case *gpb.TypedValue_JsonIetfVal:
		jsondata = val.JsonIetfVal
	case *gpb.TypedValue_JsonVal:
		jsondata = val.JsonVal
	}

	if value != nil {
		return value
	} else if jsondata != nil {
		// TODO
	}

	return nil
}

func Version() string {
	return gnmiVersion
}

func isRawRequested(output string) bool {
	return strings.HasSuffix(output, "::raw")
}

func getLabels(prefix string) (map[string]string, string) {
	labels := make(map[string]string)
	subs := labelsRegex.FindAllStringSubmatch(prefix, -1)
	for _, sub := range subs {
		if len(sub) != 3 {
			continue
		}
		kv := strings.Split(sub[2], "=")
		labels[kv[0]] = strings.ReplaceAll(kv[1], "'", "")
		prefix = strings.Replace(prefix, sub[0], sub[1], 1)
	}

	return labels, prefix
}
