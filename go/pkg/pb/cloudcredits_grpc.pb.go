// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: cloudcredits.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// CloudCreditsCreditServiceClient is the client API for CloudCreditsCreditService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CloudCreditsCreditServiceClient interface {
	// Ping always returns a successful response by the service implementation.
	// It can be used for testing connectivity to the service.
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Create(ctx context.Context, in *Credit, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ReadCredits(ctx context.Context, in *CreditFilter, opts ...grpc.CallOption) (*CreditResponse, error)
	ReadUnappliedCreditBalance(ctx context.Context, in *Account, opts ...grpc.CallOption) (*UnappliedCreditBalance, error)
	ReadInternal(ctx context.Context, in *Account, opts ...grpc.CallOption) (CloudCreditsCreditService_ReadInternalClient, error)
	CreditMigrate(ctx context.Context, in *UnappliedCredits, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DeleteMigratedCredit(ctx context.Context, in *MigratedCredits, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CreateCreditStateLog(ctx context.Context, in *CreditsState, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ReadCreditStateLog(ctx context.Context, in *CreditsStateFilter, opts ...grpc.CallOption) (*CreditsStateResponse, error)
}

type cloudCreditsCreditServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCloudCreditsCreditServiceClient(cc grpc.ClientConnInterface) CloudCreditsCreditServiceClient {
	return &cloudCreditsCreditServiceClient{cc}
}

func (c *cloudCreditsCreditServiceClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCreditService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCreditServiceClient) Create(ctx context.Context, in *Credit, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCreditService/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCreditServiceClient) ReadCredits(ctx context.Context, in *CreditFilter, opts ...grpc.CallOption) (*CreditResponse, error) {
	out := new(CreditResponse)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCreditService/ReadCredits", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCreditServiceClient) ReadUnappliedCreditBalance(ctx context.Context, in *Account, opts ...grpc.CallOption) (*UnappliedCreditBalance, error) {
	out := new(UnappliedCreditBalance)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCreditService/ReadUnappliedCreditBalance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCreditServiceClient) ReadInternal(ctx context.Context, in *Account, opts ...grpc.CallOption) (CloudCreditsCreditService_ReadInternalClient, error) {
	stream, err := c.cc.NewStream(ctx, &CloudCreditsCreditService_ServiceDesc.Streams[0], "/proto.CloudCreditsCreditService/ReadInternal", opts...)
	if err != nil {
		return nil, err
	}
	x := &cloudCreditsCreditServiceReadInternalClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type CloudCreditsCreditService_ReadInternalClient interface {
	Recv() (*Credit, error)
	grpc.ClientStream
}

type cloudCreditsCreditServiceReadInternalClient struct {
	grpc.ClientStream
}

func (x *cloudCreditsCreditServiceReadInternalClient) Recv() (*Credit, error) {
	m := new(Credit)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *cloudCreditsCreditServiceClient) CreditMigrate(ctx context.Context, in *UnappliedCredits, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCreditService/CreditMigrate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCreditServiceClient) DeleteMigratedCredit(ctx context.Context, in *MigratedCredits, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCreditService/DeleteMigratedCredit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCreditServiceClient) CreateCreditStateLog(ctx context.Context, in *CreditsState, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCreditService/CreateCreditStateLog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCreditServiceClient) ReadCreditStateLog(ctx context.Context, in *CreditsStateFilter, opts ...grpc.CallOption) (*CreditsStateResponse, error) {
	out := new(CreditsStateResponse)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCreditService/ReadCreditStateLog", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CloudCreditsCreditServiceServer is the server API for CloudCreditsCreditService service.
// All implementations must embed UnimplementedCloudCreditsCreditServiceServer
// for forward compatibility
type CloudCreditsCreditServiceServer interface {
	// Ping always returns a successful response by the service implementation.
	// It can be used for testing connectivity to the service.
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	Create(context.Context, *Credit) (*emptypb.Empty, error)
	ReadCredits(context.Context, *CreditFilter) (*CreditResponse, error)
	ReadUnappliedCreditBalance(context.Context, *Account) (*UnappliedCreditBalance, error)
	ReadInternal(*Account, CloudCreditsCreditService_ReadInternalServer) error
	CreditMigrate(context.Context, *UnappliedCredits) (*emptypb.Empty, error)
	DeleteMigratedCredit(context.Context, *MigratedCredits) (*emptypb.Empty, error)
	CreateCreditStateLog(context.Context, *CreditsState) (*emptypb.Empty, error)
	ReadCreditStateLog(context.Context, *CreditsStateFilter) (*CreditsStateResponse, error)
	mustEmbedUnimplementedCloudCreditsCreditServiceServer()
}

// UnimplementedCloudCreditsCreditServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCloudCreditsCreditServiceServer struct {
}

func (UnimplementedCloudCreditsCreditServiceServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedCloudCreditsCreditServiceServer) Create(context.Context, *Credit) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedCloudCreditsCreditServiceServer) ReadCredits(context.Context, *CreditFilter) (*CreditResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadCredits not implemented")
}
func (UnimplementedCloudCreditsCreditServiceServer) ReadUnappliedCreditBalance(context.Context, *Account) (*UnappliedCreditBalance, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadUnappliedCreditBalance not implemented")
}
func (UnimplementedCloudCreditsCreditServiceServer) ReadInternal(*Account, CloudCreditsCreditService_ReadInternalServer) error {
	return status.Errorf(codes.Unimplemented, "method ReadInternal not implemented")
}
func (UnimplementedCloudCreditsCreditServiceServer) CreditMigrate(context.Context, *UnappliedCredits) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreditMigrate not implemented")
}
func (UnimplementedCloudCreditsCreditServiceServer) DeleteMigratedCredit(context.Context, *MigratedCredits) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteMigratedCredit not implemented")
}
func (UnimplementedCloudCreditsCreditServiceServer) CreateCreditStateLog(context.Context, *CreditsState) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCreditStateLog not implemented")
}
func (UnimplementedCloudCreditsCreditServiceServer) ReadCreditStateLog(context.Context, *CreditsStateFilter) (*CreditsStateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadCreditStateLog not implemented")
}
func (UnimplementedCloudCreditsCreditServiceServer) mustEmbedUnimplementedCloudCreditsCreditServiceServer() {
}

// UnsafeCloudCreditsCreditServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CloudCreditsCreditServiceServer will
// result in compilation errors.
type UnsafeCloudCreditsCreditServiceServer interface {
	mustEmbedUnimplementedCloudCreditsCreditServiceServer()
}

func RegisterCloudCreditsCreditServiceServer(s grpc.ServiceRegistrar, srv CloudCreditsCreditServiceServer) {
	s.RegisterService(&CloudCreditsCreditService_ServiceDesc, srv)
}

func _CloudCreditsCreditService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCreditServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCreditService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCreditServiceServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCreditService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Credit)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCreditServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCreditService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCreditServiceServer).Create(ctx, req.(*Credit))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCreditService_ReadCredits_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreditFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCreditServiceServer).ReadCredits(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCreditService/ReadCredits",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCreditServiceServer).ReadCredits(ctx, req.(*CreditFilter))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCreditService_ReadUnappliedCreditBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Account)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCreditServiceServer).ReadUnappliedCreditBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCreditService/ReadUnappliedCreditBalance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCreditServiceServer).ReadUnappliedCreditBalance(ctx, req.(*Account))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCreditService_ReadInternal_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Account)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CloudCreditsCreditServiceServer).ReadInternal(m, &cloudCreditsCreditServiceReadInternalServer{stream})
}

type CloudCreditsCreditService_ReadInternalServer interface {
	Send(*Credit) error
	grpc.ServerStream
}

type cloudCreditsCreditServiceReadInternalServer struct {
	grpc.ServerStream
}

func (x *cloudCreditsCreditServiceReadInternalServer) Send(m *Credit) error {
	return x.ServerStream.SendMsg(m)
}

func _CloudCreditsCreditService_CreditMigrate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnappliedCredits)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCreditServiceServer).CreditMigrate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCreditService/CreditMigrate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCreditServiceServer).CreditMigrate(ctx, req.(*UnappliedCredits))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCreditService_DeleteMigratedCredit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MigratedCredits)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCreditServiceServer).DeleteMigratedCredit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCreditService/DeleteMigratedCredit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCreditServiceServer).DeleteMigratedCredit(ctx, req.(*MigratedCredits))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCreditService_CreateCreditStateLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreditsState)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCreditServiceServer).CreateCreditStateLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCreditService/CreateCreditStateLog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCreditServiceServer).CreateCreditStateLog(ctx, req.(*CreditsState))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCreditService_ReadCreditStateLog_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreditsStateFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCreditServiceServer).ReadCreditStateLog(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCreditService/ReadCreditStateLog",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCreditServiceServer).ReadCreditStateLog(ctx, req.(*CreditsStateFilter))
	}
	return interceptor(ctx, in, info, handler)
}

// CloudCreditsCreditService_ServiceDesc is the grpc.ServiceDesc for CloudCreditsCreditService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CloudCreditsCreditService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.CloudCreditsCreditService",
	HandlerType: (*CloudCreditsCreditServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _CloudCreditsCreditService_Ping_Handler,
		},
		{
			MethodName: "Create",
			Handler:    _CloudCreditsCreditService_Create_Handler,
		},
		{
			MethodName: "ReadCredits",
			Handler:    _CloudCreditsCreditService_ReadCredits_Handler,
		},
		{
			MethodName: "ReadUnappliedCreditBalance",
			Handler:    _CloudCreditsCreditService_ReadUnappliedCreditBalance_Handler,
		},
		{
			MethodName: "CreditMigrate",
			Handler:    _CloudCreditsCreditService_CreditMigrate_Handler,
		},
		{
			MethodName: "DeleteMigratedCredit",
			Handler:    _CloudCreditsCreditService_DeleteMigratedCredit_Handler,
		},
		{
			MethodName: "CreateCreditStateLog",
			Handler:    _CloudCreditsCreditService_CreateCreditStateLog_Handler,
		},
		{
			MethodName: "ReadCreditStateLog",
			Handler:    _CloudCreditsCreditService_ReadCreditStateLog_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ReadInternal",
			Handler:       _CloudCreditsCreditService_ReadInternal_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "cloudcredits.proto",
}

// CloudCreditsCouponServiceClient is the client API for CloudCreditsCouponService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CloudCreditsCouponServiceClient interface {
	Redeem(ctx context.Context, in *CloudCreditsCouponRedeem, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Create(ctx context.Context, in *CouponCreate, opts ...grpc.CallOption) (*Coupon, error)
	Read(ctx context.Context, in *CouponFilter, opts ...grpc.CallOption) (*CouponResponse, error)
	Disable(ctx context.Context, in *CouponDisable, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type cloudCreditsCouponServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCloudCreditsCouponServiceClient(cc grpc.ClientConnInterface) CloudCreditsCouponServiceClient {
	return &cloudCreditsCouponServiceClient{cc}
}

func (c *cloudCreditsCouponServiceClient) Redeem(ctx context.Context, in *CloudCreditsCouponRedeem, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCouponService/Redeem", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCouponServiceClient) Create(ctx context.Context, in *CouponCreate, opts ...grpc.CallOption) (*Coupon, error) {
	out := new(Coupon)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCouponService/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCouponServiceClient) Read(ctx context.Context, in *CouponFilter, opts ...grpc.CallOption) (*CouponResponse, error) {
	out := new(CouponResponse)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCouponService/Read", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cloudCreditsCouponServiceClient) Disable(ctx context.Context, in *CouponDisable, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.CloudCreditsCouponService/Disable", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CloudCreditsCouponServiceServer is the server API for CloudCreditsCouponService service.
// All implementations must embed UnimplementedCloudCreditsCouponServiceServer
// for forward compatibility
type CloudCreditsCouponServiceServer interface {
	Redeem(context.Context, *CloudCreditsCouponRedeem) (*emptypb.Empty, error)
	Create(context.Context, *CouponCreate) (*Coupon, error)
	Read(context.Context, *CouponFilter) (*CouponResponse, error)
	Disable(context.Context, *CouponDisable) (*emptypb.Empty, error)
	mustEmbedUnimplementedCloudCreditsCouponServiceServer()
}

// UnimplementedCloudCreditsCouponServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCloudCreditsCouponServiceServer struct {
}

func (UnimplementedCloudCreditsCouponServiceServer) Redeem(context.Context, *CloudCreditsCouponRedeem) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Redeem not implemented")
}
func (UnimplementedCloudCreditsCouponServiceServer) Create(context.Context, *CouponCreate) (*Coupon, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedCloudCreditsCouponServiceServer) Read(context.Context, *CouponFilter) (*CouponResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Read not implemented")
}
func (UnimplementedCloudCreditsCouponServiceServer) Disable(context.Context, *CouponDisable) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Disable not implemented")
}
func (UnimplementedCloudCreditsCouponServiceServer) mustEmbedUnimplementedCloudCreditsCouponServiceServer() {
}

// UnsafeCloudCreditsCouponServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CloudCreditsCouponServiceServer will
// result in compilation errors.
type UnsafeCloudCreditsCouponServiceServer interface {
	mustEmbedUnimplementedCloudCreditsCouponServiceServer()
}

func RegisterCloudCreditsCouponServiceServer(s grpc.ServiceRegistrar, srv CloudCreditsCouponServiceServer) {
	s.RegisterService(&CloudCreditsCouponService_ServiceDesc, srv)
}

func _CloudCreditsCouponService_Redeem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CloudCreditsCouponRedeem)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCouponServiceServer).Redeem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCouponService/Redeem",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCouponServiceServer).Redeem(ctx, req.(*CloudCreditsCouponRedeem))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCouponService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CouponCreate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCouponServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCouponService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCouponServiceServer).Create(ctx, req.(*CouponCreate))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCouponService_Read_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CouponFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCouponServiceServer).Read(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCouponService/Read",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCouponServiceServer).Read(ctx, req.(*CouponFilter))
	}
	return interceptor(ctx, in, info, handler)
}

func _CloudCreditsCouponService_Disable_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CouponDisable)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CloudCreditsCouponServiceServer).Disable(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.CloudCreditsCouponService/Disable",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CloudCreditsCouponServiceServer).Disable(ctx, req.(*CouponDisable))
	}
	return interceptor(ctx, in, info, handler)
}

// CloudCreditsCouponService_ServiceDesc is the grpc.ServiceDesc for CloudCreditsCouponService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CloudCreditsCouponService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.CloudCreditsCouponService",
	HandlerType: (*CloudCreditsCouponServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Redeem",
			Handler:    _CloudCreditsCouponService_Redeem_Handler,
		},
		{
			MethodName: "Create",
			Handler:    _CloudCreditsCouponService_Create_Handler,
		},
		{
			MethodName: "Read",
			Handler:    _CloudCreditsCouponService_Read_Handler,
		},
		{
			MethodName: "Disable",
			Handler:    _CloudCreditsCouponService_Disable_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cloudcredits.proto",
}
