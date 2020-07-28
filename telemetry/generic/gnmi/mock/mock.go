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
	return server.Send(&gnmi.SubscribeResponse{
		Response: &gnmi.SubscribeResponse_Update{
			Update: u.Notification,
		}})
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

func JuniperUpdate() *gnmi.Notification {
	//gnmi.TypedValue_AnyVal.
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
				Val:  &gnmi.TypedValue{Value: &gnmi.TypedValue_IntVal{IntVal: 23004056}},
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

//  Update:{path:{elem:{name:"__juniper_telemetry_header__"}} Val:{any_val:{type_url:"type.googleapis.com/GnmiJuniperTelemetryHeader" value:"\n\tcore1.nca\x10\xff\xff\x03\"esensor_1039_3_1:/interfaces/interface/state/counters/:/interfaces/interface/state/counters/:xmlproxyd(\x88\x80\x80\x01"}}}
//  Update:{path:{elem:{name:"__timestamp__"}} val:{uint_val:1595951912883}}
//  -update:{path:{elem:{name:"state"} elem:{name:"counters"} elem:{name:"in-octets"}} val:{uint_val:50302030597}}
//  update:{path:{elem:{name:"state"} elem:{name:"counters"} elem:{name:"in-pkts"}} val:{uint_val:23004056}}
//  update:{path:{elem:{name:"state"} elem:{name:"counters"} elem:{name:"out-octets"}} val:{uint_val:50302030597}}
//  update:{path:{elem:{name:"state"} elem:{name:"counters"} elem:{name:"out-pkts"}} val:{uint_val:23004056}}
//  update:{path:{elem:{name:"state"} elem:{name:"counters"} elem:{name:"last-clear"}} val:{string_val:"Never"}}})
