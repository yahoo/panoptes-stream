package mock

import (
	"context"
	"net"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/anypb"
)

type Response interface {
	Run(gnmi.GNMI_SubscribeServer) error
}

type GNMIServer struct {
	Resp Response
}

type Update struct {
	Notification *gnmi.Notification
	Attempt      int
}

func (*GNMIServer) Capabilities(context.Context, *gnmi.CapabilityRequest) (*gnmi.CapabilityResponse, error) {
	return nil, nil
}
func (*GNMIServer) Get(context.Context, *gnmi.GetRequest) (*gnmi.GetResponse, error) {
	return nil, nil
}
func (*GNMIServer) Set(context.Context, *gnmi.SetRequest) (*gnmi.SetResponse, error) {
	return nil, nil
}
func (g *GNMIServer) Subscribe(server gnmi.GNMI_SubscribeServer) error {
	return g.Resp.Run(server)
}

func (u Update) Run(server gnmi.GNMI_SubscribeServer) error {
	for i := 0; i < u.Attempt; i++ {
		err := server.Send(&gnmi.SubscribeResponse{
			Response: &gnmi.SubscribeResponse_Update{
				Update: u.Notification,
			}})
		if err != nil {
			return err
		}
	}

	return nil
}

func StartGNMIServer(addr string, resp Response) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	gServer := grpc.NewServer()
	mockServer := &GNMIServer{resp}
	gnmi.RegisterGNMIServer(gServer, mockServer)

	go func() {
		gServer.Serve(ln)
	}()

	return ln, nil
}

func AristaUpdate() *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: 1595363593437180059,
		Prefix:    &gnmi.Path{},
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "interfaces"},
						{Name: "interface", Key: map[string]string{"name": "Ethernet1"}},
						{Name: "state"},
						{Name: "counters"},
						{Name: "out-octets"},
					},
				},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 50302030597}},
			},
		},
	}
}

func AristaBGPUpdate() *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: 1595363593413814979,
		Prefix:    &gnmi.Path{},
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{
					Elem: []*gnmi.PathElem{
						{Name: "network-instances"},
						{Name: "network-instance", Key: map[string]string{"name": "default"}},
						{Name: "protocols"},
						{Name: "protocol", Key: map[string]string{"identifier": "BGP", "name": "BGP"}},
						{Name: "bgp"},
						{Name: "global"},
						{Name: "afi-safis"},
						{Name: "afi-safi", Key: map[string]string{"afi-safi-name": "IPV6_UNICAST"}},
						{Name: "config"},
						{Name: "afi-safi-name"},
					},
				},
				Val: &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "openconfig-bgp-types:IPV6_UNICAST"}},
			},
		},
	}
}

func CiscoXRInterface() *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: 1596928627212000000,
		Prefix: &gnmi.Path{
			Origin: "openconfig",
			Elem: []*gnmi.PathElem{
				{Name: "interfaces"},
				{Name: "interface", Key: map[string]string{"name": "GigabitEthernet0/0/0/0"}},
				{Name: "state"},
				{Name: "counters"},
			},
		},
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "in-octets"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 102387}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "out-octets"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 2642}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "in-multicast-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 65918}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "in-broadcast-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 49283}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "out-multicast-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 312}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "out-broadcast-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 9813}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "in-unknown-protos"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 76}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "in-errors"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 2}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "out-errors"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 1}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "in-unicast-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 952612}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "in-discards"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 1}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "out-unicast-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 718252}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "out-discards"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 2}},
			},
		},
	}
}

func JuniperUpdate() *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: 1595951912880990837,
		Prefix:    &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "interfaces"}, {Name: "interface", Key: map[string]string{"name": "lo0"}}}},
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "__juniper_telemetry_header__"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_AnyVal{AnyVal: &anypb.Any{TypeUrl: "type.googleapis.com/GnmiJuniperTelemetryHeader", Value: []byte("\n\tcore1.nca\x10\xff\xff\x03\"esensor_1039_3_1:/interfaces/interface/state/counters/:/interfaces/interface/state/counters/:xmlproxyd(\x88\x80\x80\x01")}}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "__timestamp__"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 1595951912883}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "in-octets"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 50302030597}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "in-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 23004050}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "out-octets"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 50302030492}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "out-pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 23004056}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "counters"}, {Name: "last-clear"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_StringVal{StringVal: "Never"}},
			},
		},
	}
}

func JuniperFakeKeyLabel() *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: 1595951912880990837,
		Prefix:    &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "interfaces"}, {Name: "interface", Key: map[string]string{"name": "lo0"}}}},
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "__juniper_telemetry_header__"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_AnyVal{AnyVal: &anypb.Any{TypeUrl: "type.googleapis.com/GnmiJuniperTelemetryHeader", Value: []byte("\n\tcore1.nca\x10\xff\xff\x03\"esensor_1039_3_1:/interfaces/interface/state/counters/:/interfaces/interface/state/counters/:xmlproxyd(\x88\x80\x80\x01")}}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "__timestamp__"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 1595951912883}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "out-queue", Key: map[string]string{"queue-number": "2"}}, {Name: "pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 50302030597}},
			},
		},
	}
}

func JuniperFakeDuplicateLabel() *gnmi.Notification {
	return &gnmi.Notification{
		Timestamp: 1595951912880990837,
		Prefix:    &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "interfaces"}, {Name: "interface", Key: map[string]string{"name": "lo0"}}}},
		Update: []*gnmi.Update{
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "__juniper_telemetry_header__"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_AnyVal{AnyVal: &anypb.Any{TypeUrl: "type.googleapis.com/GnmiJuniperTelemetryHeader", Value: []byte("\n\tcore1.nca\x10\xff\xff\x03\"esensor_1039_3_1:/interfaces/interface/state/counters/:/interfaces/interface/state/counters/:xmlproxyd(\x88\x80\x80\x01")}}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "__timestamp__"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_UintVal{UintVal: 1595951912883}},
			},
			{
				Path: &gnmi.Path{Elem: []*gnmi.PathElem{{Name: "state"}, {Name: "out-queue", Key: map[string]string{"name": "fake"}}, {Name: "pkts"}}},
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 50302030597}},
			},
		},
	}
}
