// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: training.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TrainingClusterServiceClient is the client API for TrainingClusterService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TrainingClusterServiceClient interface {
	Create(ctx context.Context, in *SlurmClusterCreateRequest, opts ...grpc.CallOption) (*SlurmClusterCreateResponse, error)
	Get(ctx context.Context, in *SlurmClusterRequest, opts ...grpc.CallOption) (*Cluster, error)
	List(ctx context.Context, in *ClusterListOption, opts ...grpc.CallOption) (*SlurmClusterResponse, error)
}

type trainingClusterServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTrainingClusterServiceClient(cc grpc.ClientConnInterface) TrainingClusterServiceClient {
	return &trainingClusterServiceClient{cc}
}

func (c *trainingClusterServiceClient) Create(ctx context.Context, in *SlurmClusterCreateRequest, opts ...grpc.CallOption) (*SlurmClusterCreateResponse, error) {
	out := new(SlurmClusterCreateResponse)
	err := c.cc.Invoke(ctx, "/proto.TrainingClusterService/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trainingClusterServiceClient) Get(ctx context.Context, in *SlurmClusterRequest, opts ...grpc.CallOption) (*Cluster, error) {
	out := new(Cluster)
	err := c.cc.Invoke(ctx, "/proto.TrainingClusterService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trainingClusterServiceClient) List(ctx context.Context, in *ClusterListOption, opts ...grpc.CallOption) (*SlurmClusterResponse, error) {
	out := new(SlurmClusterResponse)
	err := c.cc.Invoke(ctx, "/proto.TrainingClusterService/List", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TrainingClusterServiceServer is the server API for TrainingClusterService service.
// All implementations must embed UnimplementedTrainingClusterServiceServer
// for forward compatibility
type TrainingClusterServiceServer interface {
	Create(context.Context, *SlurmClusterCreateRequest) (*SlurmClusterCreateResponse, error)
	Get(context.Context, *SlurmClusterRequest) (*Cluster, error)
	List(context.Context, *ClusterListOption) (*SlurmClusterResponse, error)
	mustEmbedUnimplementedTrainingClusterServiceServer()
}

// UnimplementedTrainingClusterServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTrainingClusterServiceServer struct {
}

func (UnimplementedTrainingClusterServiceServer) Create(context.Context, *SlurmClusterCreateRequest) (*SlurmClusterCreateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedTrainingClusterServiceServer) Get(context.Context, *SlurmClusterRequest) (*Cluster, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedTrainingClusterServiceServer) List(context.Context, *ClusterListOption) (*SlurmClusterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedTrainingClusterServiceServer) mustEmbedUnimplementedTrainingClusterServiceServer() {
}

// UnsafeTrainingClusterServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TrainingClusterServiceServer will
// result in compilation errors.
type UnsafeTrainingClusterServiceServer interface {
	mustEmbedUnimplementedTrainingClusterServiceServer()
}

func RegisterTrainingClusterServiceServer(s grpc.ServiceRegistrar, srv TrainingClusterServiceServer) {
	s.RegisterService(&TrainingClusterService_ServiceDesc, srv)
}

func _TrainingClusterService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SlurmClusterCreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrainingClusterServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.TrainingClusterService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrainingClusterServiceServer).Create(ctx, req.(*SlurmClusterCreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TrainingClusterService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SlurmClusterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrainingClusterServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.TrainingClusterService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrainingClusterServiceServer).Get(ctx, req.(*SlurmClusterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TrainingClusterService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClusterListOption)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrainingClusterServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.TrainingClusterService/List",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrainingClusterServiceServer).List(ctx, req.(*ClusterListOption))
	}
	return interceptor(ctx, in, info, handler)
}

// TrainingClusterService_ServiceDesc is the grpc.ServiceDesc for TrainingClusterService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TrainingClusterService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.TrainingClusterService",
	HandlerType: (*TrainingClusterServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _TrainingClusterService_Create_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _TrainingClusterService_Get_Handler,
		},
		{
			MethodName: "List",
			Handler:    _TrainingClusterService_List_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "training.proto",
}

// TrainingBatchUserServiceClient is the client API for TrainingBatchUserService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TrainingBatchUserServiceClient interface {
	Register(ctx context.Context, in *TrainingRegistrationRequest, opts ...grpc.CallOption) (*TrainingRegistrationResponse, error)
	GetExpiryTimeById(ctx context.Context, in *GetDataRequest, opts ...grpc.CallOption) (*GetDataResponse, error)
}

type trainingBatchUserServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTrainingBatchUserServiceClient(cc grpc.ClientConnInterface) TrainingBatchUserServiceClient {
	return &trainingBatchUserServiceClient{cc}
}

func (c *trainingBatchUserServiceClient) Register(ctx context.Context, in *TrainingRegistrationRequest, opts ...grpc.CallOption) (*TrainingRegistrationResponse, error) {
	out := new(TrainingRegistrationResponse)
	err := c.cc.Invoke(ctx, "/proto.TrainingBatchUserService/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *trainingBatchUserServiceClient) GetExpiryTimeById(ctx context.Context, in *GetDataRequest, opts ...grpc.CallOption) (*GetDataResponse, error) {
	out := new(GetDataResponse)
	err := c.cc.Invoke(ctx, "/proto.TrainingBatchUserService/GetExpiryTimeById", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TrainingBatchUserServiceServer is the server API for TrainingBatchUserService service.
// All implementations must embed UnimplementedTrainingBatchUserServiceServer
// for forward compatibility
type TrainingBatchUserServiceServer interface {
	Register(context.Context, *TrainingRegistrationRequest) (*TrainingRegistrationResponse, error)
	GetExpiryTimeById(context.Context, *GetDataRequest) (*GetDataResponse, error)
	mustEmbedUnimplementedTrainingBatchUserServiceServer()
}

// UnimplementedTrainingBatchUserServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTrainingBatchUserServiceServer struct {
}

func (UnimplementedTrainingBatchUserServiceServer) Register(context.Context, *TrainingRegistrationRequest) (*TrainingRegistrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedTrainingBatchUserServiceServer) GetExpiryTimeById(context.Context, *GetDataRequest) (*GetDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetExpiryTimeById not implemented")
}
func (UnimplementedTrainingBatchUserServiceServer) mustEmbedUnimplementedTrainingBatchUserServiceServer() {
}

// UnsafeTrainingBatchUserServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TrainingBatchUserServiceServer will
// result in compilation errors.
type UnsafeTrainingBatchUserServiceServer interface {
	mustEmbedUnimplementedTrainingBatchUserServiceServer()
}

func RegisterTrainingBatchUserServiceServer(s grpc.ServiceRegistrar, srv TrainingBatchUserServiceServer) {
	s.RegisterService(&TrainingBatchUserService_ServiceDesc, srv)
}

func _TrainingBatchUserService_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TrainingRegistrationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrainingBatchUserServiceServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.TrainingBatchUserService/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrainingBatchUserServiceServer).Register(ctx, req.(*TrainingRegistrationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TrainingBatchUserService_GetExpiryTimeById_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TrainingBatchUserServiceServer).GetExpiryTimeById(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.TrainingBatchUserService/GetExpiryTimeById",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TrainingBatchUserServiceServer).GetExpiryTimeById(ctx, req.(*GetDataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TrainingBatchUserService_ServiceDesc is the grpc.ServiceDesc for TrainingBatchUserService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TrainingBatchUserService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.TrainingBatchUserService",
	HandlerType: (*TrainingBatchUserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _TrainingBatchUserService_Register_Handler,
		},
		{
			MethodName: "GetExpiryTimeById",
			Handler:    _TrainingBatchUserService_GetExpiryTimeById_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "training.proto",
}
