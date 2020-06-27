// Code generated by protoc-gen-go. DO NOT EDIT.
// source: GnmiJuniperTelemetryHeaderExtension.proto

/*
Package GnmiJuniperTelemetryHeaderExtension is a generated protocol buffer package.

It is generated from these files:
	GnmiJuniperTelemetryHeaderExtension.proto

It has these top-level messages:
	GnmiJuniperTelemetryHeaderExtension
*/
package GnmiJuniperTelemetryHeaderExtension

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type GnmiJuniperTelemetryHeaderExtension struct {
	// router name:export IP address
	SystemId string `protobuf:"bytes,1,opt,name=system_id,json=systemId" json:"system_id,omitempty"`
	// line card / RE (slot number)
	ComponentId uint32 `protobuf:"varint,2,opt,name=component_id,json=componentId" json:"component_id,omitempty"`
	// PFE (if applicable)
	SubComponentId uint32 `protobuf:"varint,3,opt,name=sub_component_id,json=subComponentId" json:"sub_component_id,omitempty"`
	// Internal sensor name
	SensorName string `protobuf:"bytes,4,opt,name=sensor_name,json=sensorName" json:"sensor_name,omitempty"`
	// Sensor path in the subscribe request
	SubscribedPath string `protobuf:"bytes,5,opt,name=subscribed_path,json=subscribedPath" json:"subscribed_path,omitempty"`
	// Internal sensor path in junos
	StreamedPath string `protobuf:"bytes,6,opt,name=streamed_path,json=streamedPath" json:"streamed_path,omitempty"`
	Component    string `protobuf:"bytes,7,opt,name=component" json:"component,omitempty"`
	// Sequence number, monotonically increasing for each
	SequenceNumber uint64 `protobuf:"varint,8,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	// Payload get timestamp in milliseconds
	PayloadGetTimestamp int64 `protobuf:"varint,9,opt,name=payload_get_timestamp,json=payloadGetTimestamp" json:"payload_get_timestamp,omitempty"`
	// Stream creation timestamp in milliseconds
	StreamCreationTimestamp int64 `protobuf:"varint,10,opt,name=stream_creation_timestamp,json=streamCreationTimestamp" json:"stream_creation_timestamp,omitempty"`
	// Event timestamp in milliseconds
	EventTimestamp int64 `protobuf:"varint,11,opt,name=event_timestamp,json=eventTimestamp" json:"event_timestamp,omitempty"`
	// Export timestamp in milliseconds
	ExportTimestamp int64 `protobuf:"varint,12,opt,name=export_timestamp,json=exportTimestamp" json:"export_timestamp,omitempty"`
	// Subsequence number
	SubSequenceNumber uint64 `protobuf:"varint,13,opt,name=sub_sequence_number,json=subSequenceNumber" json:"sub_sequence_number,omitempty"`
	// End of marker
	Eom bool `protobuf:"varint,14,opt,name=eom" json:"eom,omitempty"`
}

func (m *GnmiJuniperTelemetryHeaderExtension) Reset()         { *m = GnmiJuniperTelemetryHeaderExtension{} }
func (m *GnmiJuniperTelemetryHeaderExtension) String() string { return proto.CompactTextString(m) }
func (*GnmiJuniperTelemetryHeaderExtension) ProtoMessage()    {}
func (*GnmiJuniperTelemetryHeaderExtension) Descriptor() ([]byte, []int) {
	return fileDescriptor0, []int{0}
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetSystemId() string {
	if m != nil {
		return m.SystemId
	}
	return ""
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetComponentId() uint32 {
	if m != nil {
		return m.ComponentId
	}
	return 0
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetSubComponentId() uint32 {
	if m != nil {
		return m.SubComponentId
	}
	return 0
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetSensorName() string {
	if m != nil {
		return m.SensorName
	}
	return ""
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetSubscribedPath() string {
	if m != nil {
		return m.SubscribedPath
	}
	return ""
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetStreamedPath() string {
	if m != nil {
		return m.StreamedPath
	}
	return ""
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetComponent() string {
	if m != nil {
		return m.Component
	}
	return ""
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetSequenceNumber() uint64 {
	if m != nil {
		return m.SequenceNumber
	}
	return 0
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetPayloadGetTimestamp() int64 {
	if m != nil {
		return m.PayloadGetTimestamp
	}
	return 0
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetStreamCreationTimestamp() int64 {
	if m != nil {
		return m.StreamCreationTimestamp
	}
	return 0
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetEventTimestamp() int64 {
	if m != nil {
		return m.EventTimestamp
	}
	return 0
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetExportTimestamp() int64 {
	if m != nil {
		return m.ExportTimestamp
	}
	return 0
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetSubSequenceNumber() uint64 {
	if m != nil {
		return m.SubSequenceNumber
	}
	return 0
}

func (m *GnmiJuniperTelemetryHeaderExtension) GetEom() bool {
	if m != nil {
		return m.Eom
	}
	return false
}

func init() {
	proto.RegisterType((*GnmiJuniperTelemetryHeaderExtension)(nil), "GnmiJuniperTelemetryHeaderExtension")
}

func init() { proto.RegisterFile("GnmiJuniperTelemetryHeaderExtension.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 364 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x41, 0x4f, 0xf2, 0x40,
	0x10, 0x86, 0xd3, 0x0f, 0x3e, 0xa4, 0x03, 0x14, 0x2c, 0x31, 0xd6, 0x68, 0x62, 0x95, 0x03, 0xe5,
	0xc2, 0x41, 0x6f, 0x5e, 0x89, 0x41, 0x3c, 0x10, 0x53, 0xb9, 0x6f, 0xb6, 0xed, 0x44, 0x9a, 0xb0,
	0xbb, 0x75, 0x77, 0x6b, 0xe0, 0x37, 0xfa, 0xa7, 0x4c, 0xb7, 0x94, 0x82, 0x27, 0x6f, 0xcd, 0xf3,
	0x3e, 0x6f, 0x66, 0xa6, 0x2d, 0x4c, 0xe6, 0x9c, 0xa5, 0xaf, 0x39, 0x4f, 0x33, 0x94, 0x2b, 0xdc,
	0x20, 0x43, 0x2d, 0x77, 0x2f, 0x48, 0x13, 0x94, 0xcf, 0x5b, 0x8d, 0x5c, 0xa5, 0x82, 0x4f, 0x33,
	0x29, 0xb4, 0xb8, 0xff, 0x6e, 0xc2, 0xe8, 0x0f, 0xb6, 0x7b, 0x0d, 0xb6, 0xda, 0x29, 0x8d, 0x8c,
	0xa4, 0x89, 0x67, 0xf9, 0x56, 0x60, 0x87, 0xed, 0x12, 0x2c, 0x12, 0xf7, 0x0e, 0xba, 0xb1, 0x60,
	0x99, 0xe0, 0xc8, 0x75, 0x91, 0xff, 0xf3, 0xad, 0xa0, 0x17, 0x76, 0x0e, 0x6c, 0x91, 0xb8, 0x01,
	0x0c, 0x54, 0x1e, 0x91, 0x13, 0xad, 0x61, 0x34, 0x47, 0xe5, 0xd1, 0xec, 0xc8, 0xbc, 0x85, 0x8e,
	0x42, 0xae, 0x84, 0x24, 0x9c, 0x32, 0xf4, 0x9a, 0x66, 0x16, 0x94, 0x68, 0x49, 0x19, 0xba, 0x63,
	0xe8, 0xab, 0x3c, 0x52, 0xb1, 0x4c, 0x23, 0x4c, 0x48, 0x46, 0xf5, 0xda, 0xfb, 0x6f, 0x24, 0xa7,
	0xc6, 0x6f, 0x54, 0xaf, 0xdd, 0x11, 0xf4, 0x94, 0x96, 0x48, 0x59, 0xa5, 0xb5, 0x8c, 0xd6, 0xad,
	0xa0, 0x91, 0x6e, 0xc0, 0x3e, 0x2c, 0xe5, 0x9d, 0x19, 0xa1, 0x06, 0x66, 0x16, 0x7e, 0xe6, 0xc8,
	0x63, 0x24, 0x3c, 0x67, 0x11, 0x4a, 0xaf, 0xed, 0x5b, 0x41, 0x33, 0x74, 0x2a, 0xbc, 0x34, 0xd4,
	0x7d, 0x80, 0x8b, 0x8c, 0xee, 0x36, 0x82, 0x26, 0xe4, 0x03, 0x35, 0xd1, 0x29, 0x43, 0xa5, 0x29,
	0xcb, 0x3c, 0xdb, 0xb7, 0x82, 0x46, 0x38, 0xdc, 0x87, 0x73, 0xd4, 0xab, 0x2a, 0x72, 0x9f, 0xe0,
	0xaa, 0x5c, 0x85, 0xc4, 0x12, 0xa9, 0x4e, 0x05, 0x3f, 0xea, 0x81, 0xe9, 0x5d, 0x96, 0xc2, 0x6c,
	0x9f, 0xd7, 0xdd, 0x31, 0xf4, 0xf1, 0xab, 0x78, 0x8f, 0x75, 0xa3, 0x63, 0x1a, 0x8e, 0xc1, 0xb5,
	0x38, 0x81, 0x01, 0x6e, 0x33, 0x21, 0x8f, 0xcd, 0xae, 0x31, 0xfb, 0x25, 0xaf, 0xd5, 0x29, 0x0c,
	0x8b, 0x6f, 0xf4, 0xfb, 0xe0, 0x9e, 0x39, 0xf8, 0x5c, 0xe5, 0xd1, 0xfb, 0xe9, 0xcd, 0x03, 0x68,
	0xa0, 0x60, 0x9e, 0xe3, 0x5b, 0x41, 0x3b, 0x2c, 0x1e, 0xa3, 0x96, 0xf9, 0xa9, 0x1e, 0x7f, 0x02,
	0x00, 0x00, 0xff, 0xff, 0x14, 0xaa, 0x8f, 0xf3, 0x81, 0x02, 0x00, 0x00,
}