// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.21.4
// source: go/pkg/storage/storagecontroller/api/common.proto

package api

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

type AuthenticationContext struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Scheme:
	//
	//	*AuthenticationContext_Basic_
	//	*AuthenticationContext_Bearer_
	Scheme isAuthenticationContext_Scheme `protobuf_oneof:"scheme"`
}

func (x *AuthenticationContext) Reset() {
	*x = AuthenticationContext{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticationContext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticationContext) ProtoMessage() {}

func (x *AuthenticationContext) ProtoReflect() protoreflect.Message {
	mi := &file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticationContext.ProtoReflect.Descriptor instead.
func (*AuthenticationContext) Descriptor() ([]byte, []int) {
	return file_go_pkg_storage_storagecontroller_api_common_proto_rawDescGZIP(), []int{0}
}

func (m *AuthenticationContext) GetScheme() isAuthenticationContext_Scheme {
	if m != nil {
		return m.Scheme
	}
	return nil
}

func (x *AuthenticationContext) GetBasic() *AuthenticationContext_Basic {
	if x, ok := x.GetScheme().(*AuthenticationContext_Basic_); ok {
		return x.Basic
	}
	return nil
}

func (x *AuthenticationContext) GetBearer() *AuthenticationContext_Bearer {
	if x, ok := x.GetScheme().(*AuthenticationContext_Bearer_); ok {
		return x.Bearer
	}
	return nil
}

type isAuthenticationContext_Scheme interface {
	isAuthenticationContext_Scheme()
}

type AuthenticationContext_Basic_ struct {
	Basic *AuthenticationContext_Basic `protobuf:"bytes,1,opt,name=basic,proto3,oneof"`
}

type AuthenticationContext_Bearer_ struct {
	Bearer *AuthenticationContext_Bearer `protobuf:"bytes,2,opt,name=bearer,proto3,oneof"`
}

func (*AuthenticationContext_Basic_) isAuthenticationContext_Scheme() {}

func (*AuthenticationContext_Bearer_) isAuthenticationContext_Scheme() {}

type AuthenticationContext_Basic struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Principal   string `protobuf:"bytes,1,opt,name=principal,proto3" json:"principal,omitempty"`
	Credentials string `protobuf:"bytes,2,opt,name=credentials,proto3" json:"credentials,omitempty"`
}

func (x *AuthenticationContext_Basic) Reset() {
	*x = AuthenticationContext_Basic{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticationContext_Basic) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticationContext_Basic) ProtoMessage() {}

func (x *AuthenticationContext_Basic) ProtoReflect() protoreflect.Message {
	mi := &file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticationContext_Basic.ProtoReflect.Descriptor instead.
func (*AuthenticationContext_Basic) Descriptor() ([]byte, []int) {
	return file_go_pkg_storage_storagecontroller_api_common_proto_rawDescGZIP(), []int{0, 0}
}

func (x *AuthenticationContext_Basic) GetPrincipal() string {
	if x != nil {
		return x.Principal
	}
	return ""
}

func (x *AuthenticationContext_Basic) GetCredentials() string {
	if x != nil {
		return x.Credentials
	}
	return ""
}

type AuthenticationContext_Bearer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token string `protobuf:"bytes,2,opt,name=token,proto3" json:"token,omitempty"`
}

func (x *AuthenticationContext_Bearer) Reset() {
	*x = AuthenticationContext_Bearer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthenticationContext_Bearer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthenticationContext_Bearer) ProtoMessage() {}

func (x *AuthenticationContext_Bearer) ProtoReflect() protoreflect.Message {
	mi := &file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthenticationContext_Bearer.ProtoReflect.Descriptor instead.
func (*AuthenticationContext_Bearer) Descriptor() ([]byte, []int) {
	return file_go_pkg_storage_storagecontroller_api_common_proto_rawDescGZIP(), []int{0, 1}
}

func (x *AuthenticationContext_Bearer) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

var File_go_pkg_storage_storagecontroller_api_common_proto protoreflect.FileDescriptor

var file_go_pkg_storage_storagecontroller_api_common_proto_rawDesc = []byte{
	0x0a, 0x31, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c,
	0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x1a, 0x69, 0x6e, 0x74, 0x65, 0x6c, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x22,
	0xaf, 0x02, 0x0a, 0x15, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x12, 0x4f, 0x0a, 0x05, 0x62, 0x61, 0x73,
	0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x37, 0x2e, 0x69, 0x6e, 0x74, 0x65, 0x6c,
	0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c,
	0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x2e, 0x42, 0x61, 0x73, 0x69,
	0x63, 0x48, 0x00, 0x52, 0x05, 0x62, 0x61, 0x73, 0x69, 0x63, 0x12, 0x52, 0x0a, 0x06, 0x62, 0x65,
	0x61, 0x72, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x38, 0x2e, 0x69, 0x6e, 0x74,
	0x65, 0x6c, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f,
	0x6c, 0x6c, 0x65, 0x72, 0x2e, 0x76, 0x31, 0x2e, 0x41, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x78, 0x74, 0x2e, 0x42, 0x65,
	0x61, 0x72, 0x65, 0x72, 0x48, 0x00, 0x52, 0x06, 0x62, 0x65, 0x61, 0x72, 0x65, 0x72, 0x1a, 0x47,
	0x0a, 0x05, 0x42, 0x61, 0x73, 0x69, 0x63, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72, 0x69, 0x6e, 0x63,
	0x69, 0x70, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x69, 0x6e,
	0x63, 0x69, 0x70, 0x61, 0x6c, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x61, 0x6c, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x63, 0x72, 0x65, 0x64,
	0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x1a, 0x1e, 0x0a, 0x06, 0x42, 0x65, 0x61, 0x72, 0x65,
	0x72, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x42, 0x08, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d,
	0x65, 0x42, 0x6a, 0x5a, 0x68, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x69, 0x6e, 0x74, 0x65, 0x6c, 0x2d, 0x69, 0x6e, 0x6e, 0x65, 0x72, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x2f, 0x66, 0x72, 0x61, 0x6d, 0x65, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x64, 0x65, 0x76, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x2e, 0x69, 0x64, 0x63, 0x2f, 0x67, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f,
	0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x63,
	0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x2f, 0x61, 0x70, 0x69, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_go_pkg_storage_storagecontroller_api_common_proto_rawDescOnce sync.Once
	file_go_pkg_storage_storagecontroller_api_common_proto_rawDescData = file_go_pkg_storage_storagecontroller_api_common_proto_rawDesc
)

func file_go_pkg_storage_storagecontroller_api_common_proto_rawDescGZIP() []byte {
	file_go_pkg_storage_storagecontroller_api_common_proto_rawDescOnce.Do(func() {
		file_go_pkg_storage_storagecontroller_api_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_go_pkg_storage_storagecontroller_api_common_proto_rawDescData)
	})
	return file_go_pkg_storage_storagecontroller_api_common_proto_rawDescData
}

var file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_go_pkg_storage_storagecontroller_api_common_proto_goTypes = []interface{}{
	(*AuthenticationContext)(nil),        // 0: intel.storagecontroller.v1.AuthenticationContext
	(*AuthenticationContext_Basic)(nil),  // 1: intel.storagecontroller.v1.AuthenticationContext.Basic
	(*AuthenticationContext_Bearer)(nil), // 2: intel.storagecontroller.v1.AuthenticationContext.Bearer
}
var file_go_pkg_storage_storagecontroller_api_common_proto_depIdxs = []int32{
	1, // 0: intel.storagecontroller.v1.AuthenticationContext.basic:type_name -> intel.storagecontroller.v1.AuthenticationContext.Basic
	2, // 1: intel.storagecontroller.v1.AuthenticationContext.bearer:type_name -> intel.storagecontroller.v1.AuthenticationContext.Bearer
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_go_pkg_storage_storagecontroller_api_common_proto_init() }
func file_go_pkg_storage_storagecontroller_api_common_proto_init() {
	if File_go_pkg_storage_storagecontroller_api_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthenticationContext); i {
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
		file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthenticationContext_Basic); i {
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
		file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthenticationContext_Bearer); i {
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
	file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*AuthenticationContext_Basic_)(nil),
		(*AuthenticationContext_Bearer_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_go_pkg_storage_storagecontroller_api_common_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_go_pkg_storage_storagecontroller_api_common_proto_goTypes,
		DependencyIndexes: file_go_pkg_storage_storagecontroller_api_common_proto_depIdxs,
		MessageInfos:      file_go_pkg_storage_storagecontroller_api_common_proto_msgTypes,
	}.Build()
	File_go_pkg_storage_storagecontroller_api_common_proto = out.File
	file_go_pkg_storage_storagecontroller_api_common_proto_rawDesc = nil
	file_go_pkg_storage_storagecontroller_api_common_proto_goTypes = nil
	file_go_pkg_storage_storagecontroller_api_common_proto_depIdxs = nil
}
