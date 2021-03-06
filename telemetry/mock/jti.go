//: Copyright Verizon Media
//: Licensed under the terms of the Apache 2.0 License. See LICENSE file in the project root for terms.

package mock

import (
	"context"
	"net"

	"google.golang.org/grpc"

	jpb "github.com/yahoo/panoptes-stream/telemetry/juniper/proto/telemetry"
)

// JTIResponse represents mock JTI response
type JTIResponse interface {
	Run(*jpb.SubscriptionRequest, jpb.OpenConfigTelemetry_TelemetrySubscribeServer) error
}

// OpenConfigTelemetryServer represents OpenConfigTelemetry mock Server
type OpenConfigTelemetryServer struct {
	Resp JTIResponse
}

// TelemetrySubscribe is mock TelemetrySubscribe method
func (oc *OpenConfigTelemetryServer) TelemetrySubscribe(subReq *jpb.SubscriptionRequest, subServer jpb.OpenConfigTelemetry_TelemetrySubscribeServer) error {
	return oc.Resp.Run(subReq, subServer)
}

// CancelTelemetrySubscription is mock CancelTelemetrySubscription method
func (*OpenConfigTelemetryServer) CancelTelemetrySubscription(context.Context, *jpb.CancelSubscriptionRequest) (*jpb.CancelSubscriptionReply, error) {
	return nil, nil
}

// GetTelemetrySubscriptions is mock GetTelemetrySubscriptions method
func (*OpenConfigTelemetryServer) GetTelemetrySubscriptions(context.Context, *jpb.GetSubscriptionsRequest) (*jpb.GetSubscriptionsReply, error) {
	return nil, nil
}

// GetTelemetryOperationalState is mock GetTelemetryOperationalState method
func (*OpenConfigTelemetryServer) GetTelemetryOperationalState(context.Context, *jpb.GetOperationalStateRequest) (*jpb.GetOperationalStateReply, error) {
	return nil, nil
}

// GetDataEncodings is mock GetDataEncodings method
func (*OpenConfigTelemetryServer) GetDataEncodings(context.Context, *jpb.DataEncodingRequest) (*jpb.DataEncodingReply, error) {
	return nil, nil
}

// StartJTIServer starts JTI mock server
func StartJTIServer(addr string, resp JTIResponse) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	gServer := grpc.NewServer()
	mockServer := &OpenConfigTelemetryServer{resp}
	jpb.RegisterOpenConfigTelemetryServer(gServer, mockServer)

	go func() {
		gServer.Serve(ln)
	}()

	return ln, nil
}

// JuniperJTILo0InterfaceSample returns JTI mock data included an interface metrics
func JuniperJTILo0InterfaceSample() *jpb.OpenConfigData {
	return &jpb.OpenConfigData{
		SystemId:       "core1.lax",
		ComponentId:    65535,
		Path:           "sensor_1040_3_1:/interfaces/interface[name='lo0']/state/counters/:/interfaces/interface[name='lo0']/state/counters/:xmlproxyd",
		Timestamp:      1596067993610,
		SequenceNumber: 1,
		Kv: []*jpb.KeyValue{
			{Key: "__timestamp__", Value: &jpb.KeyValue_UintValue{UintValue: 1596067993611}},
			{Key: "__junos_re_stream_creation_timestamp__", Value: &jpb.KeyValue_UintValue{UintValue: 1596067993562}},
			{Key: "__junos_re_payload_get_timestamp__", Value: &jpb.KeyValue_UintValue{UintValue: 1596067993609}},
			{Key: "__prefix__", Value: &jpb.KeyValue_StrValue{StrValue: "/interfaces/interface[name='lo0']/"}},
			{Key: "state/counters/in-octets", Value: &jpb.KeyValue_UintValue{UintValue: 52613105736}},
			{Key: "state/counters/in-pkts", Value: &jpb.KeyValue_UintValue{UintValue: 23609955}},
			{Key: "state/counters/out-octets", Value: &jpb.KeyValue_UintValue{UintValue: 52613105736}},
			{Key: "state/counters/out-pkts", Value: &jpb.KeyValue_UintValue{UintValue: 23609955}},
			{Key: "state/counters/last-clear", Value: &jpb.KeyValue_StrValue{StrValue: "Never"}},
		},
	}
}

// JuniperJTIMix returns JTI mock data included fake metrics
func JuniperJTIMix() *jpb.OpenConfigData {
	return &jpb.OpenConfigData{
		SystemId:       "core1.lax",
		ComponentId:    65535,
		Path:           "sensor_1040_3_1:/mixes/mix[name='lo0']/state/:/mixes/mix[name='lo0']/state/:xmlproxyd",
		Timestamp:      1596067993610,
		SequenceNumber: 1,
		Kv: []*jpb.KeyValue{
			{Key: "__timestamp__", Value: &jpb.KeyValue_UintValue{UintValue: 1596067993611}},
			{Key: "__junos_re_stream_creation_timestamp__", Value: &jpb.KeyValue_UintValue{UintValue: 1596067993562}},
			{Key: "__junos_re_payload_get_timestamp__", Value: &jpb.KeyValue_UintValue{UintValue: 1596067993609}},
			{Key: "__prefix__", Value: &jpb.KeyValue_StrValue{StrValue: "/interfaces/interface[name='lo0']/"}},
			{Key: "state/counters/in-octets", Value: &jpb.KeyValue_UintValue{UintValue: 52613105736}},
			{Key: "state/counters/out-queue[queue-number=0]/pkts", Value: &jpb.KeyValue_UintValue{UintValue: 526131}},
		},
	}
}

// JuniperBGPSample returns JTI mock data included BGP metrics
func JuniperBGPSample() *jpb.OpenConfigData {
	return &jpb.OpenConfigData{
		SystemId:       "core1.lax",
		ComponentId:    65535,
		Path:           "sensor_1040:/network-instances/network-instance/protocols/protocol/bgp/:/network-instances/network-instance/protocols/protocol/bgp/:rpd",
		Timestamp:      1596087032354,
		SequenceNumber: 1,
		Kv: []*jpb.KeyValue{
			{Key: "__timestamp__", Value: &jpb.KeyValue_UintValue{UintValue: 1596087032354}},
			{Key: "__junos_re_stream_creation_timestamp__", Value: &jpb.KeyValue_UintValue{UintValue: 1596087032023}},
			{Key: "__junos_re_payload_get_timestamp__", Value: &jpb.KeyValue_UintValue{UintValue: 1596087032329}},
			{Key: "__prefix__", Value: &jpb.KeyValue_StrValue{StrValue: "/network-instances/network-instance[instance-name='master']/protocols/protocol/bgp/peer-groups/peer-group[peer-group-name='BUR']/afi-safis/afi-safi[afi-safi-name='IPV4_UNICAST']/"}},
			{Key: "graceful-restart/state/enabled", Value: &jpb.KeyValue_BoolValue{BoolValue: true}},
			{Key: "state/afi-safi-name", Value: &jpb.KeyValue_StrValue{StrValue: "IPV4_UNICAST"}},
			{Key: "state/enabled", Value: &jpb.KeyValue_BoolValue{BoolValue: true}},
			{Key: "state/prefix-limit/state/max-prefixes", Value: &jpb.KeyValue_UintValue{UintValue: 0}},
			{Key: "state/prefix-limit/state/shutdown-threshold-pct", Value: &jpb.KeyValue_UintValue{UintValue: 0}},
			{Key: "add-paths/receive", Value: &jpb.KeyValue_BoolValue{BoolValue: false}},
			{Key: "add-paths/send-max", Value: &jpb.KeyValue_UintValue{UintValue: 0}},
			{Key: "add-paths/eligible-prefix-policy", Value: &jpb.KeyValue_StrValue{StrValue: ""}},
			{Key: "__prefix__", Value: &jpb.KeyValue_StrValue{StrValue: "/network-instances/network-instance[instance-name='master']/protocols/protocol/bgp/peer-groups/peer-group[peer-group-name='BUR']/"}},
			{Key: "graceful-restart/state/enabled", Value: &jpb.KeyValue_BoolValue{BoolValue: false}},
			{Key: "graceful-restart/state/restart-time", Value: &jpb.KeyValue_UintValue{UintValue: 120}},
			{Key: "graceful-restart/state/helper-only", Value: &jpb.KeyValue_BoolValue{BoolValue: true}},
			{Key: "apply-policy/state/import-policy", Value: &jpb.KeyValue_StrValue{StrValue: "IMPORT_FROM_BUR"}},
			{Key: "apply-policy/state/import-policy", Value: &jpb.KeyValue_StrValue{StrValue: "NOTHING"}},
			{Key: "apply-policy/state/export-policy", Value: &jpb.KeyValue_StrValue{StrValue: "EXPORT_TO_BUR"}},
			{Key: "apply-policy/state/export-policy", Value: &jpb.KeyValue_StrValue{StrValue: "NOTHING"}},
			{Key: "__prefix__", Value: &jpb.KeyValue_StrValue{StrValue: "/network-instances/network-instance[instance-name='master']/protocols/protocol/bgp/peer-groups/peer-group[peer-group-name='BURv6']/"}},
			{Key: "state/peer-as", Value: &jpb.KeyValue_UintValue{UintValue: 65531}},
			{Key: "state/local-as", Value: &jpb.KeyValue_UintValue{UintValue: 65400}},
			{Key: "state/peer-type", Value: &jpb.KeyValue_StrValue{StrValue: "EXTERNAL"}},
			{Key: "state/auth-password", Value: &jpb.KeyValue_StrValue{StrValue: "(null)"}},
			{Key: "state/remove-private-as", Value: &jpb.KeyValue_StrValue{StrValue: "255"}},
			{Key: "state/route-flap-damping", Value: &jpb.KeyValue_BoolValue{BoolValue: false}},
			{Key: "state/description", Value: &jpb.KeyValue_StrValue{StrValue: "core1.lax"}},
			{Key: "state/peer-group-name", Value: &jpb.KeyValue_StrValue{StrValue: "BURv6"}},
			{Key: "state/total-paths", Value: &jpb.KeyValue_UintValue{UintValue: 88243}},
			{Key: "state/total-prefixes", Value: &jpb.KeyValue_UintValue{UintValue: 88243}},
			{Key: "transport/state/tcp-mss", Value: &jpb.KeyValue_UintValue{UintValue: 0}},
			{Key: "transport/state/mtu-discovery", Value: &jpb.KeyValue_BoolValue{BoolValue: false}},
			{Key: "transport/state/passive-mode", Value: &jpb.KeyValue_BoolValue{BoolValue: false}},
		},
	}
}
