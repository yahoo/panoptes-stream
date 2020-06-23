package gnmi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"juniper/proto/GnmiJuniperTelemetryHeader"

	"git.vzbuilders.com/marshadrad/panoptes/telemetry"

	"github.com/golang/protobuf/ptypes"
	"github.com/openconfig/gnmi/path"
	"github.com/openconfig/ygot/ygot"
	"google.golang.org/grpc"

	apb "github.com/golang/protobuf/ptypes/any"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	_ "github.com/openconfig/gnmi/proto/gnmi_ext"
)

func Register() {
	telemetry.Register(nil)
}

type KeyValues map[string]interface{}

type gnmi struct {
	conn          *grpc.ClientConn
	client        gpb.GNMIClient
	subscriptions []*gpb.Subscription

	dataChan chan *gpb.SubscribeResponse
}

func NewGNMI(cfg *config) *gnmi {
	var subscriptions []*gpb.Subscription
	for _, sensor := range cfg.Sensors {
		if sensor.Type == "gnmi" {
			path, _ := ygot.StringToPath(sensor.Path, ygot.StructuredPath, ygot.StringSlicePath)

			mode := gpb.SubscriptionMode_value[strings.ToUpper(sensor.Mode)]
			sampleInterval, _ := time.ParseDuration("10s")
			subscriptions = append(subscriptions, &gpb.Subscription{
				Path:              path,
				Mode:              gpb.SubscriptionMode(mode),
				SampleInterval:    uint64(sampleInterval.Nanoseconds()),
				SuppressRedundant: false,
			})
		}
	}
	return &gnmi{
		subscriptions: subscriptions,
		dataChan:      make(chan *gpb.SubscribeResponse, 100),
	}
}

func (g *gnmi) run(ctx context.Context) {
	g.client = gpb.NewGNMIClient(g.conn)
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

	subClient, err := g.client.Subscribe(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = subClient.Send(subReq)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1; i++ {
		go g.worker(ctx)
	}

	for ctx.Err() == nil {
		var resp *gpb.SubscribeResponse
		if resp, err = subClient.Recv(); err != nil {
			log.Println(err)
			break
		}

		g.dataChan <- resp
	}

}
func (g *gnmi) worker(ctx context.Context) {
	for {
		select {
		case d, ok := <-g.dataChan:
			if !ok {
				return
			}

			//TODO ext := d.GetExtension()

			switch resp := d.Response.(type) {
			case *gpb.SubscribeResponse_Update:
				kv := g.decoder(resp)
				kv.PrettyPrint()
			case *gpb.SubscribeResponse_SyncResponse:
			case *gpb.SubscribeResponse_Error:
			}
		case <-ctx.Done():
			return
		}
	}
}

func (g *gnmi) decoder(resp *gpb.SubscribeResponse_Update) KeyValues {
	kv := make(KeyValues)
	kv["__service__"] = "gnmi_v0.7.0"

	kv["__update_timestamp__"] = resp.Update.GetTimestamp()
	kv["__prefix__"] = path.ToStrings(resp.Update.GetPrefix(), true)

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
				log.Println("ERR:", err)
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
			kv[key] = value
		} else if jsondata != nil {
			// TODO
		}

	}

	return kv
}

func (kv KeyValues) PrettyPrint() error {
	b, err := json.MarshalIndent(kv, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}

func isGNMIInterfacesSensor(p []string) bool {
	if len(p) < 3 {
		return false
	}

	if p[0] == "interfaces" && p[1] == "interface" {
		return true
	}

	return false
}

func getGNMIIfNameFromPrefix(p []string) string {
	if len(p) == 3 {
		return p[2]
	}

	if len(p) == 6 {
		return fmt.Sprintf("%s.%s", p[2], p[5])
	}

	return "unknown"
}
