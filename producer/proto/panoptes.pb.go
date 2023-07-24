// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.4
// source: panoptes.proto

package panoptes

import (
	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Panoptes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SystemId  string            `protobuf:"bytes,1,opt,name=system_id,json=systemId,proto3" json:"system_id,omitempty"`
	Prefix    string            `protobuf:"bytes,2,opt,name=prefix,proto3" json:"prefix,omitempty"`
	Labels    map[string]string `protobuf:"bytes,3,rep,name=labels,proto3" json:"labels,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Timestamp int64             `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Key       string            `protobuf:"bytes,5,opt,name=key,proto3" json:"key,omitempty"`
	Value     *any.Any          `protobuf:"bytes,6,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Panoptes) Reset() {
	*x = Panoptes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_panoptes_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Panoptes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Panoptes) ProtoMessage() {}

func (x *Panoptes) ProtoReflect() protoreflect.Message {
	mi := &file_panoptes_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Panoptes.ProtoReflect.Descriptor instead.
func (*Panoptes) Descriptor() ([]byte, []int) {
	return file_panoptes_proto_rawDescGZIP(), []int{0}
}

func (x *Panoptes) GetSystemId() string {
	if x != nil {
		return x.SystemId
	}
	return ""
}

func (x *Panoptes) GetPrefix() string {
	if x != nil {
		return x.Prefix
	}
	return ""
}

func (x *Panoptes) GetLabels() map[string]string {
	if x != nil {
		return x.Labels
	}
	return nil
}

func (x *Panoptes) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *Panoptes) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Panoptes) GetValue() *any.Any {
	if x != nil {
		return x.Value
	}
	return nil
}

var File_panoptes_proto protoreflect.FileDescriptor

var file_panoptes_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x70, 0x61, 0x6e, 0x6f, 0x70, 0x74, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x70, 0x61, 0x6e, 0x6f, 0x70, 0x74, 0x65, 0x73, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8e, 0x02, 0x0a, 0x08, 0x70, 0x61, 0x6e, 0x6f, 0x70, 0x74,
	0x65, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x5f, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x49, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x70, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x36, 0x0a, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c,
	0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x70, 0x61, 0x6e, 0x6f, 0x70, 0x74,
	0x65, 0x73, 0x2e, 0x70, 0x61, 0x6e, 0x6f, 0x70, 0x74, 0x65, 0x73, 0x2e, 0x4c, 0x61, 0x62, 0x65,
	0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x73, 0x12,
	0x1c, 0x0a, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x2a, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x41, 0x6e, 0x79, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x1a, 0x39, 0x0a, 0x0b, 0x4c,
	0x61, 0x62, 0x65, 0x6c, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_panoptes_proto_rawDescOnce sync.Once
	file_panoptes_proto_rawDescData = file_panoptes_proto_rawDesc
)

func file_panoptes_proto_rawDescGZIP() []byte {
	file_panoptes_proto_rawDescOnce.Do(func() {
		file_panoptes_proto_rawDescData = protoimpl.X.CompressGZIP(file_panoptes_proto_rawDescData)
	})
	return file_panoptes_proto_rawDescData
}

var file_panoptes_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_panoptes_proto_goTypes = []interface{}{
	(*Panoptes)(nil), // 0: panoptes.panoptes
	nil,              // 1: panoptes.panoptes.LabelsEntry
	(*any.Any)(nil),  // 2: google.protobuf.Any
}
var file_panoptes_proto_depIdxs = []int32{
	1, // 0: panoptes.panoptes.labels:type_name -> panoptes.panoptes.LabelsEntry
	2, // 1: panoptes.panoptes.value:type_name -> google.protobuf.Any
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_panoptes_proto_init() }
func file_panoptes_proto_init() {
	if File_panoptes_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_panoptes_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Panoptes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_panoptes_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_panoptes_proto_goTypes,
		DependencyIndexes: file_panoptes_proto_depIdxs,
		MessageInfos:      file_panoptes_proto_msgTypes,
	}.Build()
	File_panoptes_proto = out.File
	file_panoptes_proto_rawDesc = nil
	file_panoptes_proto_goTypes = nil
	file_panoptes_proto_depIdxs = nil
}