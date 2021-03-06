//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package mock

import (
	"net"

	telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"

	mdtGRPC "github.com/yahoo/panoptes-stream/telemetry/cisco/proto"
)

// MDTServer represents MDT mock server.
type MDTServer struct{}

// CreateSubs returns MDT gRPC response.
func (m *MDTServer) CreateSubs(sub *mdtGRPC.SubscribeRequest, g mdtGRPC.GRPCConfigOper_CreateSubsServer) error {
	tm := MDTInterface()
	b, err := proto.Marshal(tm)
	if err != nil {
		return err
	}

	subResp := &mdtGRPC.SubscribeResponse{
		RequestId: 1,
		Data:      b,
	}

	return g.Send(subResp)
}

// MDTInterface returns Cisco MDT mock data included two recursive sets.
func MDTInterface() *telemetry.Telemetry {
	return &telemetry.Telemetry{
		NodeId:              &telemetry.Telemetry_NodeIdStr{NodeIdStr: "ios"},
		Subscription:        &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "Sub3"},
		EncodingPath:        "openconfig-interfaces:interfaces/interface",
		CollectionId:        8,
		CollectionStartTime: 1597098790358,
		MsgTimestamp:        1597098790358,
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Timestamp: 1597098790450,
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "GigabitEthernet0/0/0/0"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name: "hold-time",
								Fields: []*telemetry.TelemetryField{
									{
										Name: "state",
										Fields: []*telemetry.TelemetryField{
											{
												Name:        "up",
												ValueByType: &telemetry.TelemetryField_Uint32Value{Uint32Value: 10},
											},
											{
												Name:        "down",
												ValueByType: &telemetry.TelemetryField_Uint32Value{Uint32Value: 0},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Timestamp: 1597098790480,
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "GigabitEthernet0/0/0/1"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name: "hold-time",
								Fields: []*telemetry.TelemetryField{
									{
										Name: "state",
										Fields: []*telemetry.TelemetryField{
											{
												Name:        "up",
												ValueByType: &telemetry.TelemetryField_Uint32Value{Uint32Value: 10},
											},
											{
												Name:        "down",
												ValueByType: &telemetry.TelemetryField_Uint32Value{Uint32Value: 0},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		CollectionEndTime: 1597098790873,
	}

}

// MDTInterfaceII returns Cisco MDT mock data included two recursive sets (interface counters).
func MDTInterfaceII() *telemetry.Telemetry {
	return &telemetry.Telemetry{
		NodeId:              &telemetry.Telemetry_NodeIdStr{NodeIdStr: "ios"},
		Subscription:        &telemetry.Telemetry_SubscriptionIdStr{SubscriptionIdStr: "Sub3"},
		EncodingPath:        "openconfig-interfaces:interfaces/interface",
		CollectionId:        11,
		CollectionStartTime: 1597098790977,
		MsgTimestamp:        1597098790977,
		DataGpbkv: []*telemetry.TelemetryField{
			{
				Timestamp: 1597098791076,
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "GigabitEthernet0/0/0/0"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name: "state",
								Fields: []*telemetry.TelemetryField{
									{
										Name: "counters",
										Fields: []*telemetry.TelemetryField{
											{
												Name:        "in-octets",
												ValueByType: &telemetry.TelemetryField_Uint64Value{Uint64Value: 1023},
											},
											{
												Name:        "out-octets",
												ValueByType: &telemetry.TelemetryField_Uint64Value{Uint64Value: 872},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Timestamp: 1597098791086,
				Fields: []*telemetry.TelemetryField{
					{
						Name: "keys",
						Fields: []*telemetry.TelemetryField{
							{
								Name:        "name",
								ValueByType: &telemetry.TelemetryField_StringValue{StringValue: "GigabitEthernet0/0/0/1"},
							},
						},
					},
					{
						Name: "content",
						Fields: []*telemetry.TelemetryField{
							{
								Name: "state",
								Fields: []*telemetry.TelemetryField{
									{
										Name: "counters",
										Fields: []*telemetry.TelemetryField{
											{
												Name:        "in-octets",
												ValueByType: &telemetry.TelemetryField_Uint64Value{Uint64Value: 1223},
											},
											{
												Name:        "out-octets",
												ValueByType: &telemetry.TelemetryField_Uint64Value{Uint64Value: 8172},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		CollectionEndTime: 1597098791977,
	}
}

// StartMDTServer starts MDT mock server.
func StartMDTServer(addr string) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	gServer := grpc.NewServer()
	mockServer := &MDTServer{}
	mdtGRPC.RegisterGRPCConfigOperServer(gServer, mockServer)

	go func() {
		gServer.Serve(ln)
	}()

	return ln, nil
}
