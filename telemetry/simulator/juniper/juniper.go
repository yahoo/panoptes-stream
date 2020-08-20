package juniper

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/protobuf/types/known/anypb"
)

type ifce struct {
	name string
	m    int64

	inOcteds  int64
	inPkts    int64
	outOcteds int64
	outPkts   int64
}

// Update represents gNMI notification / update
type Update struct {
	Notification *gnmi.Notification
}

func (i *ifce) update() {
	i.inOcteds += rand.Int63n(10000 * i.m)
	i.inPkts += rand.Int63n(100 * i.m)
	i.outOcteds += rand.Int63n(10000 * i.m)
	i.outPkts += rand.Int63n(500 * i.m)
}

// New constructs juniper simpulator update
func New() *Update {
	return &Update{}
}

// Run sends gNMI updates
func (u Update) Run(server gnmi.GNMI_SubscribeServer) error {
	ifces := []*ifce{}
	for pic := 0; pic < 2; pic++ {
		rnd := rand.Int63n(5) + 1
		for port := 0; port < 6; port++ {
			for channel := 0; channel < 4; channel++ {
				ifces = append(ifces, &ifce{
					name: fmt.Sprintf("xe-0/%d/%d:%d", pic, port, channel), m: rnd})
			}
		}
	}

	for {
		for _, ifc := range ifces {
			ifc.update()

			err := server.Send(&gnmi.SubscribeResponse{
				Response: &gnmi.SubscribeResponse_Update{
					Update: juniperUpdate(ifc),
				}})
			if err != nil {
				return err
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func juniperUpdate(i *ifce) *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: time.Now().UnixNano(),
		Prefix:    &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "interfaces"}, {Name: "interface", Key: map[string]string{"name": i.name}}}},
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "__juniper_telemetry_header__"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_AnyVal{AnyVal: &anypb.Any{TypeUrl: "type.googleapis.com/GnmiJuniperTelemetryHeader", Value: []byte("\n\tcore1.nca\x10\xff\xff\x03\"esensor_1039_3_1:/interfaces/interface/state/counters/:/interfaces/interface/state/counters/:xmlproxyd(\x88\x80\x80\x01")}}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "__timestamp__"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: uint64(time.Now().UnixNano() / 1000)}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "in-octets"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: i.inOcteds}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "in-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: i.inPkts}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "out-octets"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: i.outOcteds}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "out-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: i.outPkts}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "last-clear"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "Never"}},
			},
		},
	}
}
