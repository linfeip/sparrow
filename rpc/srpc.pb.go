// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.0
// source: srpc.proto

package rpc

import (
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

type CallType int32

const (
	CallType_Request      CallType = 0
	CallType_Response     CallType = 1
	CallType_BidiStream   CallType = 2
	CallType_ClientStream CallType = 3
	CallType_ServerStream CallType = 4
	CallType_StreamClosed CallType = 5
)

// Enum value maps for CallType.
var (
	CallType_name = map[int32]string{
		0: "Request",
		1: "Response",
		2: "BidiStream",
		3: "ClientStream",
		4: "ServerStream",
		5: "StreamClosed",
	}
	CallType_value = map[string]int32{
		"Request":      0,
		"Response":     1,
		"BidiStream":   2,
		"ClientStream": 3,
		"ServerStream": 4,
		"StreamClosed": 5,
	}
)

func (x CallType) Enum() *CallType {
	p := new(CallType)
	*p = x
	return p
}

func (x CallType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (CallType) Descriptor() protoreflect.EnumDescriptor {
	return file_srpc_proto_enumTypes[0].Descriptor()
}

func (CallType) Type() protoreflect.EnumType {
	return &file_srpc_proto_enumTypes[0]
}

func (x CallType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use CallType.Descriptor instead.
func (CallType) EnumDescriptor() ([]byte, []int) {
	return file_srpc_proto_rawDescGZIP(), []int{0}
}

type ProtoPayload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type          CallType    `protobuf:"varint,1,opt,name=type,proto3,enum=rpc.CallType" json:"type,omitempty"`
	StreamId      uint64      `protobuf:"varint,2,opt,name=streamId,proto3" json:"streamId,omitempty"`
	Route         string      `protobuf:"bytes,3,opt,name=route,proto3" json:"route,omitempty"`
	SerializeType int32       `protobuf:"varint,4,opt,name=serializeType,proto3" json:"serializeType,omitempty"`
	Data          []byte      `protobuf:"bytes,5,opt,name=data,proto3" json:"data,omitempty"`
	Error         *ProtoError `protobuf:"bytes,6,opt,name=error,proto3" json:"error,omitempty"`
	Header        []*Header   `protobuf:"bytes,7,rep,name=header,proto3" json:"header,omitempty"`
}

func (x *ProtoPayload) Reset() {
	*x = ProtoPayload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_srpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoPayload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoPayload) ProtoMessage() {}

func (x *ProtoPayload) ProtoReflect() protoreflect.Message {
	mi := &file_srpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoPayload.ProtoReflect.Descriptor instead.
func (*ProtoPayload) Descriptor() ([]byte, []int) {
	return file_srpc_proto_rawDescGZIP(), []int{0}
}

func (x *ProtoPayload) GetType() CallType {
	if x != nil {
		return x.Type
	}
	return CallType_Request
}

func (x *ProtoPayload) GetStreamId() uint64 {
	if x != nil {
		return x.StreamId
	}
	return 0
}

func (x *ProtoPayload) GetRoute() string {
	if x != nil {
		return x.Route
	}
	return ""
}

func (x *ProtoPayload) GetSerializeType() int32 {
	if x != nil {
		return x.SerializeType
	}
	return 0
}

func (x *ProtoPayload) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *ProtoPayload) GetError() *ProtoError {
	if x != nil {
		return x.Error
	}
	return nil
}

func (x *ProtoPayload) GetHeader() []*Header {
	if x != nil {
		return x.Header
	}
	return nil
}

type ProtoError struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ErrCode int32  `protobuf:"varint,6,opt,name=errCode,proto3" json:"errCode,omitempty"`
	ErrMsg  string `protobuf:"bytes,7,opt,name=errMsg,proto3" json:"errMsg,omitempty"`
}

func (x *ProtoError) Reset() {
	*x = ProtoError{}
	if protoimpl.UnsafeEnabled {
		mi := &file_srpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ProtoError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProtoError) ProtoMessage() {}

func (x *ProtoError) ProtoReflect() protoreflect.Message {
	mi := &file_srpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ProtoError.ProtoReflect.Descriptor instead.
func (*ProtoError) Descriptor() ([]byte, []int) {
	return file_srpc_proto_rawDescGZIP(), []int{1}
}

func (x *ProtoError) GetErrCode() int32 {
	if x != nil {
		return x.ErrCode
	}
	return 0
}

func (x *ProtoError) GetErrMsg() string {
	if x != nil {
		return x.ErrMsg
	}
	return ""
}

type Header struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Header) Reset() {
	*x = Header{}
	if protoimpl.UnsafeEnabled {
		mi := &file_srpc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_srpc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_srpc_proto_rawDescGZIP(), []int{2}
}

func (x *Header) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Header) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_srpc_proto protoreflect.FileDescriptor

var file_srpc_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x72, 0x70,
	0x63, 0x22, 0xe9, 0x01, 0x0a, 0x0c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x12, 0x21, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x0d, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x61, 0x6c, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x49,
	0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x73, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x49,
	0x64, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x73, 0x65, 0x72, 0x69, 0x61,
	0x6c, 0x69, 0x7a, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d,
	0x73, 0x65, 0x72, 0x69, 0x61, 0x6c, 0x69, 0x7a, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x12, 0x25, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0f, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x23, 0x0a, 0x06, 0x68, 0x65, 0x61, 0x64,
	0x65, 0x72, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x48,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x22, 0x3e, 0x0a,
	0x0a, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x65,
	0x72, 0x72, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x65, 0x72,
	0x72, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x72, 0x72, 0x4d, 0x73, 0x67, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x65, 0x72, 0x72, 0x4d, 0x73, 0x67, 0x22, 0x30, 0x0a,
	0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x2a,
	0x6b, 0x0a, 0x08, 0x43, 0x61, 0x6c, 0x6c, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0b, 0x0a, 0x07, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x42, 0x69, 0x64, 0x69, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x10, 0x02, 0x12, 0x10, 0x0a, 0x0c, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x10, 0x03, 0x12, 0x10, 0x0a, 0x0c, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x10, 0x04, 0x12, 0x10, 0x0a, 0x0c, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x43, 0x6c, 0x6f, 0x73, 0x65, 0x64, 0x10, 0x05, 0x42, 0x08, 0x5a, 0x06,
	0x2e, 0x2f, 0x3b, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_srpc_proto_rawDescOnce sync.Once
	file_srpc_proto_rawDescData = file_srpc_proto_rawDesc
)

func file_srpc_proto_rawDescGZIP() []byte {
	file_srpc_proto_rawDescOnce.Do(func() {
		file_srpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_srpc_proto_rawDescData)
	})
	return file_srpc_proto_rawDescData
}

var file_srpc_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_srpc_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_srpc_proto_goTypes = []interface{}{
	(CallType)(0),        // 0: rpc.CallType
	(*ProtoPayload)(nil), // 1: rpc.ProtoPayload
	(*ProtoError)(nil),   // 2: rpc.ProtoError
	(*Header)(nil),       // 3: rpc.Header
}
var file_srpc_proto_depIdxs = []int32{
	0, // 0: rpc.ProtoPayload.type:type_name -> rpc.CallType
	2, // 1: rpc.ProtoPayload.error:type_name -> rpc.ProtoError
	3, // 2: rpc.ProtoPayload.header:type_name -> rpc.Header
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_srpc_proto_init() }
func file_srpc_proto_init() {
	if File_srpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_srpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoPayload); i {
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
		file_srpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ProtoError); i {
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
		file_srpc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Header); i {
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
			RawDescriptor: file_srpc_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_srpc_proto_goTypes,
		DependencyIndexes: file_srpc_proto_depIdxs,
		EnumInfos:         file_srpc_proto_enumTypes,
		MessageInfos:      file_srpc_proto_msgTypes,
	}.Build()
	File_srpc_proto = out.File
	file_srpc_proto_rawDesc = nil
	file_srpc_proto_goTypes = nil
	file_srpc_proto_depIdxs = nil
}
