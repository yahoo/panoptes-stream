package mock

import telemetry "github.com/cisco-ie/nx-telemetry-proto/telemetry_bis"

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
