// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: infaas-dispatcher.proto

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

// DispatcherClient is the client API for Dispatcher service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DispatcherClient interface {
	GenerateStream(ctx context.Context, in *DispatcherRequest, opts ...grpc.CallOption) (Dispatcher_GenerateStreamClient, error)
	Generate(ctx context.Context, in *DispatcherRequest, opts ...grpc.CallOption) (*DispatcherResponse, error)
	// DoWork is the agent entry point
	DoWork(ctx context.Context, opts ...grpc.CallOption) (Dispatcher_DoWorkClient, error)
}

type dispatcherClient struct {
	cc grpc.ClientConnInterface
}

func NewDispatcherClient(cc grpc.ClientConnInterface) DispatcherClient {
	return &dispatcherClient{cc}
}

func (c *dispatcherClient) GenerateStream(ctx context.Context, in *DispatcherRequest, opts ...grpc.CallOption) (Dispatcher_GenerateStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Dispatcher_ServiceDesc.Streams[0], "/proto.Dispatcher/GenerateStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &dispatcherGenerateStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Dispatcher_GenerateStreamClient interface {
	Recv() (*DispatcherResponse, error)
	grpc.ClientStream
}

type dispatcherGenerateStreamClient struct {
	grpc.ClientStream
}

func (x *dispatcherGenerateStreamClient) Recv() (*DispatcherResponse, error) {
	m := new(DispatcherResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *dispatcherClient) Generate(ctx context.Context, in *DispatcherRequest, opts ...grpc.CallOption) (*DispatcherResponse, error) {
	out := new(DispatcherResponse)
	err := c.cc.Invoke(ctx, "/proto.Dispatcher/Generate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dispatcherClient) DoWork(ctx context.Context, opts ...grpc.CallOption) (Dispatcher_DoWorkClient, error) {
	stream, err := c.cc.NewStream(ctx, &Dispatcher_ServiceDesc.Streams[1], "/proto.Dispatcher/DoWork", opts...)
	if err != nil {
		return nil, err
	}
	x := &dispatcherDoWorkClient{stream}
	return x, nil
}

type Dispatcher_DoWorkClient interface {
	Send(*DispatcherResponse) error
	Recv() (*DispatcherRequest, error)
	grpc.ClientStream
}

type dispatcherDoWorkClient struct {
	grpc.ClientStream
}

func (x *dispatcherDoWorkClient) Send(m *DispatcherResponse) error {
	return x.ClientStream.SendMsg(m)
}

func (x *dispatcherDoWorkClient) Recv() (*DispatcherRequest, error) {
	m := new(DispatcherRequest)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DispatcherServer is the server API for Dispatcher service.
// All implementations must embed UnimplementedDispatcherServer
// for forward compatibility
type DispatcherServer interface {
	GenerateStream(*DispatcherRequest, Dispatcher_GenerateStreamServer) error
	Generate(context.Context, *DispatcherRequest) (*DispatcherResponse, error)
	// DoWork is the agent entry point
	DoWork(Dispatcher_DoWorkServer) error
	mustEmbedUnimplementedDispatcherServer()
}

// UnimplementedDispatcherServer must be embedded to have forward compatible implementations.
type UnimplementedDispatcherServer struct {
}

func (UnimplementedDispatcherServer) GenerateStream(*DispatcherRequest, Dispatcher_GenerateStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method GenerateStream not implemented")
}
func (UnimplementedDispatcherServer) Generate(context.Context, *DispatcherRequest) (*DispatcherResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Generate not implemented")
}
func (UnimplementedDispatcherServer) DoWork(Dispatcher_DoWorkServer) error {
	return status.Errorf(codes.Unimplemented, "method DoWork not implemented")
}
func (UnimplementedDispatcherServer) mustEmbedUnimplementedDispatcherServer() {}

// UnsafeDispatcherServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DispatcherServer will
// result in compilation errors.
type UnsafeDispatcherServer interface {
	mustEmbedUnimplementedDispatcherServer()
}

func RegisterDispatcherServer(s grpc.ServiceRegistrar, srv DispatcherServer) {
	s.RegisterService(&Dispatcher_ServiceDesc, srv)
}

func _Dispatcher_GenerateStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(DispatcherRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DispatcherServer).GenerateStream(m, &dispatcherGenerateStreamServer{stream})
}

type Dispatcher_GenerateStreamServer interface {
	Send(*DispatcherResponse) error
	grpc.ServerStream
}

type dispatcherGenerateStreamServer struct {
	grpc.ServerStream
}

func (x *dispatcherGenerateStreamServer) Send(m *DispatcherResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _Dispatcher_Generate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DispatcherRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DispatcherServer).Generate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Dispatcher/Generate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DispatcherServer).Generate(ctx, req.(*DispatcherRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Dispatcher_DoWork_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DispatcherServer).DoWork(&dispatcherDoWorkServer{stream})
}

type Dispatcher_DoWorkServer interface {
	Send(*DispatcherRequest) error
	Recv() (*DispatcherResponse, error)
	grpc.ServerStream
}

type dispatcherDoWorkServer struct {
	grpc.ServerStream
}

func (x *dispatcherDoWorkServer) Send(m *DispatcherRequest) error {
	return x.ServerStream.SendMsg(m)
}

func (x *dispatcherDoWorkServer) Recv() (*DispatcherResponse, error) {
	m := new(DispatcherResponse)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Dispatcher_ServiceDesc is the grpc.ServiceDesc for Dispatcher service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Dispatcher_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Dispatcher",
	HandlerType: (*DispatcherServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Generate",
			Handler:    _Dispatcher_Generate_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GenerateStream",
			Handler:       _Dispatcher_GenerateStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "DoWork",
			Handler:       _Dispatcher_DoWork_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "infaas-dispatcher.proto",
}
