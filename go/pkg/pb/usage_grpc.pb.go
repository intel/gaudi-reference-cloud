// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.9
// source: usage.proto

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

// UsageServiceClient is the client API for UsageService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UsageServiceClient interface {
	PostBulkUploadResourceUsages(ctx context.Context, in *BulkUploadResourceUsages, opts ...grpc.CallOption) (*BulkUploadResourceUsagesFailed, error)
	CreateResourceUsage(ctx context.Context, in *ResourceUsageCreate, opts ...grpc.CallOption) (*ResourceUsage, error)
	UpdateResourceUsage(ctx context.Context, in *ResourceUsageUpdate, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetResourceUsageById(ctx context.Context, in *ResourceUsageId, opts ...grpc.CallOption) (*ResourceUsage, error)
	MarkResourceUsageAsReported(ctx context.Context, in *ResourceUsageId, opts ...grpc.CallOption) (*emptypb.Empty, error)
	GetProductUsageById(ctx context.Context, in *ProductUsageId, opts ...grpc.CallOption) (*ProductUsage, error)
	SearchResourceUsages(ctx context.Context, in *ResourceUsagesFilter, opts ...grpc.CallOption) (*ResourceUsages, error)
	SearchProductUsages(ctx context.Context, in *ProductUsagesFilter, opts ...grpc.CallOption) (*ProductUsages, error)
	UpdateProductUsageReport(ctx context.Context, in *ReportProductUsageUpdate, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SearchProductUsagesReport(ctx context.Context, in *ProductUsagesReportFilter, opts ...grpc.CallOption) (UsageService_SearchProductUsagesReportClient, error)
	MarkProductUsageAsReported(ctx context.Context, in *ReportProductUsageId, opts ...grpc.CallOption) (*emptypb.Empty, error)
	// defaults to returning usages for products.
	SearchUsages(ctx context.Context, in *UsagesFilter, opts ...grpc.CallOption) (*Usages, error)
	// UsageService server-side streaming return a stream of ResourceUsages messages
	StreamSearchResourceUsages(ctx context.Context, in *ResourceUsagesFilter, opts ...grpc.CallOption) (UsageService_StreamSearchResourceUsagesClient, error)
	// UsageService server-side streaming return a stream of ProductUsages messages
	StreamSearchProductUsages(ctx context.Context, in *ProductUsagesFilter, opts ...grpc.CallOption) (UsageService_StreamSearchProductUsagesClient, error)
	// Ping always returns a successful response by the service implementation.
	// It can be used for testing connectivity to the service.
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type usageServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUsageServiceClient(cc grpc.ClientConnInterface) UsageServiceClient {
	return &usageServiceClient{cc}
}

func (c *usageServiceClient) PostBulkUploadResourceUsages(ctx context.Context, in *BulkUploadResourceUsages, opts ...grpc.CallOption) (*BulkUploadResourceUsagesFailed, error) {
	out := new(BulkUploadResourceUsagesFailed)
	err := c.cc.Invoke(ctx, "/proto.UsageService/PostBulkUploadResourceUsages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) CreateResourceUsage(ctx context.Context, in *ResourceUsageCreate, opts ...grpc.CallOption) (*ResourceUsage, error) {
	out := new(ResourceUsage)
	err := c.cc.Invoke(ctx, "/proto.UsageService/CreateResourceUsage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) UpdateResourceUsage(ctx context.Context, in *ResourceUsageUpdate, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.UsageService/UpdateResourceUsage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) GetResourceUsageById(ctx context.Context, in *ResourceUsageId, opts ...grpc.CallOption) (*ResourceUsage, error) {
	out := new(ResourceUsage)
	err := c.cc.Invoke(ctx, "/proto.UsageService/GetResourceUsageById", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) MarkResourceUsageAsReported(ctx context.Context, in *ResourceUsageId, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.UsageService/MarkResourceUsageAsReported", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) GetProductUsageById(ctx context.Context, in *ProductUsageId, opts ...grpc.CallOption) (*ProductUsage, error) {
	out := new(ProductUsage)
	err := c.cc.Invoke(ctx, "/proto.UsageService/GetProductUsageById", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) SearchResourceUsages(ctx context.Context, in *ResourceUsagesFilter, opts ...grpc.CallOption) (*ResourceUsages, error) {
	out := new(ResourceUsages)
	err := c.cc.Invoke(ctx, "/proto.UsageService/SearchResourceUsages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) SearchProductUsages(ctx context.Context, in *ProductUsagesFilter, opts ...grpc.CallOption) (*ProductUsages, error) {
	out := new(ProductUsages)
	err := c.cc.Invoke(ctx, "/proto.UsageService/SearchProductUsages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) UpdateProductUsageReport(ctx context.Context, in *ReportProductUsageUpdate, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.UsageService/UpdateProductUsageReport", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) SearchProductUsagesReport(ctx context.Context, in *ProductUsagesReportFilter, opts ...grpc.CallOption) (UsageService_SearchProductUsagesReportClient, error) {
	stream, err := c.cc.NewStream(ctx, &UsageService_ServiceDesc.Streams[0], "/proto.UsageService/SearchProductUsagesReport", opts...)
	if err != nil {
		return nil, err
	}
	x := &usageServiceSearchProductUsagesReportClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type UsageService_SearchProductUsagesReportClient interface {
	Recv() (*ReportProductUsage, error)
	grpc.ClientStream
}

type usageServiceSearchProductUsagesReportClient struct {
	grpc.ClientStream
}

func (x *usageServiceSearchProductUsagesReportClient) Recv() (*ReportProductUsage, error) {
	m := new(ReportProductUsage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *usageServiceClient) MarkProductUsageAsReported(ctx context.Context, in *ReportProductUsageId, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.UsageService/MarkProductUsageAsReported", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) SearchUsages(ctx context.Context, in *UsagesFilter, opts ...grpc.CallOption) (*Usages, error) {
	out := new(Usages)
	err := c.cc.Invoke(ctx, "/proto.UsageService/SearchUsages", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageServiceClient) StreamSearchResourceUsages(ctx context.Context, in *ResourceUsagesFilter, opts ...grpc.CallOption) (UsageService_StreamSearchResourceUsagesClient, error) {
	stream, err := c.cc.NewStream(ctx, &UsageService_ServiceDesc.Streams[1], "/proto.UsageService/StreamSearchResourceUsages", opts...)
	if err != nil {
		return nil, err
	}
	x := &usageServiceStreamSearchResourceUsagesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type UsageService_StreamSearchResourceUsagesClient interface {
	Recv() (*ResourceUsage, error)
	grpc.ClientStream
}

type usageServiceStreamSearchResourceUsagesClient struct {
	grpc.ClientStream
}

func (x *usageServiceStreamSearchResourceUsagesClient) Recv() (*ResourceUsage, error) {
	m := new(ResourceUsage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *usageServiceClient) StreamSearchProductUsages(ctx context.Context, in *ProductUsagesFilter, opts ...grpc.CallOption) (UsageService_StreamSearchProductUsagesClient, error) {
	stream, err := c.cc.NewStream(ctx, &UsageService_ServiceDesc.Streams[2], "/proto.UsageService/StreamSearchProductUsages", opts...)
	if err != nil {
		return nil, err
	}
	x := &usageServiceStreamSearchProductUsagesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type UsageService_StreamSearchProductUsagesClient interface {
	Recv() (*ProductUsage, error)
	grpc.ClientStream
}

type usageServiceStreamSearchProductUsagesClient struct {
	grpc.ClientStream
}

func (x *usageServiceStreamSearchProductUsagesClient) Recv() (*ProductUsage, error) {
	m := new(ProductUsage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *usageServiceClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.UsageService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UsageServiceServer is the server API for UsageService service.
// All implementations must embed UnimplementedUsageServiceServer
// for forward compatibility
type UsageServiceServer interface {
	PostBulkUploadResourceUsages(context.Context, *BulkUploadResourceUsages) (*BulkUploadResourceUsagesFailed, error)
	CreateResourceUsage(context.Context, *ResourceUsageCreate) (*ResourceUsage, error)
	UpdateResourceUsage(context.Context, *ResourceUsageUpdate) (*emptypb.Empty, error)
	GetResourceUsageById(context.Context, *ResourceUsageId) (*ResourceUsage, error)
	MarkResourceUsageAsReported(context.Context, *ResourceUsageId) (*emptypb.Empty, error)
	GetProductUsageById(context.Context, *ProductUsageId) (*ProductUsage, error)
	SearchResourceUsages(context.Context, *ResourceUsagesFilter) (*ResourceUsages, error)
	SearchProductUsages(context.Context, *ProductUsagesFilter) (*ProductUsages, error)
	UpdateProductUsageReport(context.Context, *ReportProductUsageUpdate) (*emptypb.Empty, error)
	SearchProductUsagesReport(*ProductUsagesReportFilter, UsageService_SearchProductUsagesReportServer) error
	MarkProductUsageAsReported(context.Context, *ReportProductUsageId) (*emptypb.Empty, error)
	// defaults to returning usages for products.
	SearchUsages(context.Context, *UsagesFilter) (*Usages, error)
	// UsageService server-side streaming return a stream of ResourceUsages messages
	StreamSearchResourceUsages(*ResourceUsagesFilter, UsageService_StreamSearchResourceUsagesServer) error
	// UsageService server-side streaming return a stream of ProductUsages messages
	StreamSearchProductUsages(*ProductUsagesFilter, UsageService_StreamSearchProductUsagesServer) error
	// Ping always returns a successful response by the service implementation.
	// It can be used for testing connectivity to the service.
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedUsageServiceServer()
}

// UnimplementedUsageServiceServer must be embedded to have forward compatible implementations.
type UnimplementedUsageServiceServer struct {
}

func (UnimplementedUsageServiceServer) PostBulkUploadResourceUsages(context.Context, *BulkUploadResourceUsages) (*BulkUploadResourceUsagesFailed, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostBulkUploadResourceUsages not implemented")
}
func (UnimplementedUsageServiceServer) CreateResourceUsage(context.Context, *ResourceUsageCreate) (*ResourceUsage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateResourceUsage not implemented")
}
func (UnimplementedUsageServiceServer) UpdateResourceUsage(context.Context, *ResourceUsageUpdate) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateResourceUsage not implemented")
}
func (UnimplementedUsageServiceServer) GetResourceUsageById(context.Context, *ResourceUsageId) (*ResourceUsage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetResourceUsageById not implemented")
}
func (UnimplementedUsageServiceServer) MarkResourceUsageAsReported(context.Context, *ResourceUsageId) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarkResourceUsageAsReported not implemented")
}
func (UnimplementedUsageServiceServer) GetProductUsageById(context.Context, *ProductUsageId) (*ProductUsage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProductUsageById not implemented")
}
func (UnimplementedUsageServiceServer) SearchResourceUsages(context.Context, *ResourceUsagesFilter) (*ResourceUsages, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchResourceUsages not implemented")
}
func (UnimplementedUsageServiceServer) SearchProductUsages(context.Context, *ProductUsagesFilter) (*ProductUsages, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchProductUsages not implemented")
}
func (UnimplementedUsageServiceServer) UpdateProductUsageReport(context.Context, *ReportProductUsageUpdate) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateProductUsageReport not implemented")
}
func (UnimplementedUsageServiceServer) SearchProductUsagesReport(*ProductUsagesReportFilter, UsageService_SearchProductUsagesReportServer) error {
	return status.Errorf(codes.Unimplemented, "method SearchProductUsagesReport not implemented")
}
func (UnimplementedUsageServiceServer) MarkProductUsageAsReported(context.Context, *ReportProductUsageId) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarkProductUsageAsReported not implemented")
}
func (UnimplementedUsageServiceServer) SearchUsages(context.Context, *UsagesFilter) (*Usages, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchUsages not implemented")
}
func (UnimplementedUsageServiceServer) StreamSearchResourceUsages(*ResourceUsagesFilter, UsageService_StreamSearchResourceUsagesServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamSearchResourceUsages not implemented")
}
func (UnimplementedUsageServiceServer) StreamSearchProductUsages(*ProductUsagesFilter, UsageService_StreamSearchProductUsagesServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamSearchProductUsages not implemented")
}
func (UnimplementedUsageServiceServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedUsageServiceServer) mustEmbedUnimplementedUsageServiceServer() {}

// UnsafeUsageServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UsageServiceServer will
// result in compilation errors.
type UnsafeUsageServiceServer interface {
	mustEmbedUnimplementedUsageServiceServer()
}

func RegisterUsageServiceServer(s grpc.ServiceRegistrar, srv UsageServiceServer) {
	s.RegisterService(&UsageService_ServiceDesc, srv)
}

func _UsageService_PostBulkUploadResourceUsages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BulkUploadResourceUsages)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).PostBulkUploadResourceUsages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/PostBulkUploadResourceUsages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).PostBulkUploadResourceUsages(ctx, req.(*BulkUploadResourceUsages))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_CreateResourceUsage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceUsageCreate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).CreateResourceUsage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/CreateResourceUsage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).CreateResourceUsage(ctx, req.(*ResourceUsageCreate))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_UpdateResourceUsage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceUsageUpdate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).UpdateResourceUsage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/UpdateResourceUsage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).UpdateResourceUsage(ctx, req.(*ResourceUsageUpdate))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_GetResourceUsageById_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceUsageId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).GetResourceUsageById(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/GetResourceUsageById",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).GetResourceUsageById(ctx, req.(*ResourceUsageId))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_MarkResourceUsageAsReported_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceUsageId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).MarkResourceUsageAsReported(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/MarkResourceUsageAsReported",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).MarkResourceUsageAsReported(ctx, req.(*ResourceUsageId))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_GetProductUsageById_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProductUsageId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).GetProductUsageById(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/GetProductUsageById",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).GetProductUsageById(ctx, req.(*ProductUsageId))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_SearchResourceUsages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResourceUsagesFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).SearchResourceUsages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/SearchResourceUsages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).SearchResourceUsages(ctx, req.(*ResourceUsagesFilter))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_SearchProductUsages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProductUsagesFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).SearchProductUsages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/SearchProductUsages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).SearchProductUsages(ctx, req.(*ProductUsagesFilter))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_UpdateProductUsageReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportProductUsageUpdate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).UpdateProductUsageReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/UpdateProductUsageReport",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).UpdateProductUsageReport(ctx, req.(*ReportProductUsageUpdate))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_SearchProductUsagesReport_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ProductUsagesReportFilter)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UsageServiceServer).SearchProductUsagesReport(m, &usageServiceSearchProductUsagesReportServer{stream})
}

type UsageService_SearchProductUsagesReportServer interface {
	Send(*ReportProductUsage) error
	grpc.ServerStream
}

type usageServiceSearchProductUsagesReportServer struct {
	grpc.ServerStream
}

func (x *usageServiceSearchProductUsagesReportServer) Send(m *ReportProductUsage) error {
	return x.ServerStream.SendMsg(m)
}

func _UsageService_MarkProductUsageAsReported_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportProductUsageId)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).MarkProductUsageAsReported(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/MarkProductUsageAsReported",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).MarkProductUsageAsReported(ctx, req.(*ReportProductUsageId))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_SearchUsages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UsagesFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).SearchUsages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/SearchUsages",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).SearchUsages(ctx, req.(*UsagesFilter))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageService_StreamSearchResourceUsages_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ResourceUsagesFilter)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UsageServiceServer).StreamSearchResourceUsages(m, &usageServiceStreamSearchResourceUsagesServer{stream})
}

type UsageService_StreamSearchResourceUsagesServer interface {
	Send(*ResourceUsage) error
	grpc.ServerStream
}

type usageServiceStreamSearchResourceUsagesServer struct {
	grpc.ServerStream
}

func (x *usageServiceStreamSearchResourceUsagesServer) Send(m *ResourceUsage) error {
	return x.ServerStream.SendMsg(m)
}

func _UsageService_StreamSearchProductUsages_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ProductUsagesFilter)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UsageServiceServer).StreamSearchProductUsages(m, &usageServiceStreamSearchProductUsagesServer{stream})
}

type UsageService_StreamSearchProductUsagesServer interface {
	Send(*ProductUsage) error
	grpc.ServerStream
}

type usageServiceStreamSearchProductUsagesServer struct {
	grpc.ServerStream
}

func (x *usageServiceStreamSearchProductUsagesServer) Send(m *ProductUsage) error {
	return x.ServerStream.SendMsg(m)
}

func _UsageService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageServiceServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// UsageService_ServiceDesc is the grpc.ServiceDesc for UsageService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UsageService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.UsageService",
	HandlerType: (*UsageServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PostBulkUploadResourceUsages",
			Handler:    _UsageService_PostBulkUploadResourceUsages_Handler,
		},
		{
			MethodName: "CreateResourceUsage",
			Handler:    _UsageService_CreateResourceUsage_Handler,
		},
		{
			MethodName: "UpdateResourceUsage",
			Handler:    _UsageService_UpdateResourceUsage_Handler,
		},
		{
			MethodName: "GetResourceUsageById",
			Handler:    _UsageService_GetResourceUsageById_Handler,
		},
		{
			MethodName: "MarkResourceUsageAsReported",
			Handler:    _UsageService_MarkResourceUsageAsReported_Handler,
		},
		{
			MethodName: "GetProductUsageById",
			Handler:    _UsageService_GetProductUsageById_Handler,
		},
		{
			MethodName: "SearchResourceUsages",
			Handler:    _UsageService_SearchResourceUsages_Handler,
		},
		{
			MethodName: "SearchProductUsages",
			Handler:    _UsageService_SearchProductUsages_Handler,
		},
		{
			MethodName: "UpdateProductUsageReport",
			Handler:    _UsageService_UpdateProductUsageReport_Handler,
		},
		{
			MethodName: "MarkProductUsageAsReported",
			Handler:    _UsageService_MarkProductUsageAsReported_Handler,
		},
		{
			MethodName: "SearchUsages",
			Handler:    _UsageService_SearchUsages_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _UsageService_Ping_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SearchProductUsagesReport",
			Handler:       _UsageService_SearchProductUsagesReport_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StreamSearchResourceUsages",
			Handler:       _UsageService_StreamSearchResourceUsages_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StreamSearchProductUsages",
			Handler:       _UsageService_StreamSearchProductUsages_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "usage.proto",
}

// UsageRecordServiceClient is the client API for UsageRecordService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UsageRecordServiceClient interface {
	CreateProductUsageRecord(ctx context.Context, in *ProductUsageRecordCreate, opts ...grpc.CallOption) (*emptypb.Empty, error)
	SearchProductUsageRecords(ctx context.Context, in *ProductUsageRecordsFilter, opts ...grpc.CallOption) (UsageRecordService_SearchProductUsageRecordsClient, error)
	SearchInvalidProductUsageRecords(ctx context.Context, in *InvalidProductUsageRecordsFilter, opts ...grpc.CallOption) (UsageRecordService_SearchInvalidProductUsageRecordsClient, error)
	// Ping always returns a successful response by the service implementation.
	// It can be used for testing connectivity to the service.
	Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type usageRecordServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUsageRecordServiceClient(cc grpc.ClientConnInterface) UsageRecordServiceClient {
	return &usageRecordServiceClient{cc}
}

func (c *usageRecordServiceClient) CreateProductUsageRecord(ctx context.Context, in *ProductUsageRecordCreate, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.UsageRecordService/CreateProductUsageRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *usageRecordServiceClient) SearchProductUsageRecords(ctx context.Context, in *ProductUsageRecordsFilter, opts ...grpc.CallOption) (UsageRecordService_SearchProductUsageRecordsClient, error) {
	stream, err := c.cc.NewStream(ctx, &UsageRecordService_ServiceDesc.Streams[0], "/proto.UsageRecordService/SearchProductUsageRecords", opts...)
	if err != nil {
		return nil, err
	}
	x := &usageRecordServiceSearchProductUsageRecordsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type UsageRecordService_SearchProductUsageRecordsClient interface {
	Recv() (*ProductUsageRecord, error)
	grpc.ClientStream
}

type usageRecordServiceSearchProductUsageRecordsClient struct {
	grpc.ClientStream
}

func (x *usageRecordServiceSearchProductUsageRecordsClient) Recv() (*ProductUsageRecord, error) {
	m := new(ProductUsageRecord)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *usageRecordServiceClient) SearchInvalidProductUsageRecords(ctx context.Context, in *InvalidProductUsageRecordsFilter, opts ...grpc.CallOption) (UsageRecordService_SearchInvalidProductUsageRecordsClient, error) {
	stream, err := c.cc.NewStream(ctx, &UsageRecordService_ServiceDesc.Streams[1], "/proto.UsageRecordService/SearchInvalidProductUsageRecords", opts...)
	if err != nil {
		return nil, err
	}
	x := &usageRecordServiceSearchInvalidProductUsageRecordsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type UsageRecordService_SearchInvalidProductUsageRecordsClient interface {
	Recv() (*InvalidProductUsageRecord, error)
	grpc.ClientStream
}

type usageRecordServiceSearchInvalidProductUsageRecordsClient struct {
	grpc.ClientStream
}

func (x *usageRecordServiceSearchInvalidProductUsageRecordsClient) Recv() (*InvalidProductUsageRecord, error) {
	m := new(InvalidProductUsageRecord)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *usageRecordServiceClient) Ping(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/proto.UsageRecordService/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UsageRecordServiceServer is the server API for UsageRecordService service.
// All implementations must embed UnimplementedUsageRecordServiceServer
// for forward compatibility
type UsageRecordServiceServer interface {
	CreateProductUsageRecord(context.Context, *ProductUsageRecordCreate) (*emptypb.Empty, error)
	SearchProductUsageRecords(*ProductUsageRecordsFilter, UsageRecordService_SearchProductUsageRecordsServer) error
	SearchInvalidProductUsageRecords(*InvalidProductUsageRecordsFilter, UsageRecordService_SearchInvalidProductUsageRecordsServer) error
	// Ping always returns a successful response by the service implementation.
	// It can be used for testing connectivity to the service.
	Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error)
	mustEmbedUnimplementedUsageRecordServiceServer()
}

// UnimplementedUsageRecordServiceServer must be embedded to have forward compatible implementations.
type UnimplementedUsageRecordServiceServer struct {
}

func (UnimplementedUsageRecordServiceServer) CreateProductUsageRecord(context.Context, *ProductUsageRecordCreate) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateProductUsageRecord not implemented")
}
func (UnimplementedUsageRecordServiceServer) SearchProductUsageRecords(*ProductUsageRecordsFilter, UsageRecordService_SearchProductUsageRecordsServer) error {
	return status.Errorf(codes.Unimplemented, "method SearchProductUsageRecords not implemented")
}
func (UnimplementedUsageRecordServiceServer) SearchInvalidProductUsageRecords(*InvalidProductUsageRecordsFilter, UsageRecordService_SearchInvalidProductUsageRecordsServer) error {
	return status.Errorf(codes.Unimplemented, "method SearchInvalidProductUsageRecords not implemented")
}
func (UnimplementedUsageRecordServiceServer) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedUsageRecordServiceServer) mustEmbedUnimplementedUsageRecordServiceServer() {}

// UnsafeUsageRecordServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UsageRecordServiceServer will
// result in compilation errors.
type UnsafeUsageRecordServiceServer interface {
	mustEmbedUnimplementedUsageRecordServiceServer()
}

func RegisterUsageRecordServiceServer(s grpc.ServiceRegistrar, srv UsageRecordServiceServer) {
	s.RegisterService(&UsageRecordService_ServiceDesc, srv)
}

func _UsageRecordService_CreateProductUsageRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProductUsageRecordCreate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageRecordServiceServer).CreateProductUsageRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageRecordService/CreateProductUsageRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageRecordServiceServer).CreateProductUsageRecord(ctx, req.(*ProductUsageRecordCreate))
	}
	return interceptor(ctx, in, info, handler)
}

func _UsageRecordService_SearchProductUsageRecords_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ProductUsageRecordsFilter)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UsageRecordServiceServer).SearchProductUsageRecords(m, &usageRecordServiceSearchProductUsageRecordsServer{stream})
}

type UsageRecordService_SearchProductUsageRecordsServer interface {
	Send(*ProductUsageRecord) error
	grpc.ServerStream
}

type usageRecordServiceSearchProductUsageRecordsServer struct {
	grpc.ServerStream
}

func (x *usageRecordServiceSearchProductUsageRecordsServer) Send(m *ProductUsageRecord) error {
	return x.ServerStream.SendMsg(m)
}

func _UsageRecordService_SearchInvalidProductUsageRecords_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(InvalidProductUsageRecordsFilter)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UsageRecordServiceServer).SearchInvalidProductUsageRecords(m, &usageRecordServiceSearchInvalidProductUsageRecordsServer{stream})
}

type UsageRecordService_SearchInvalidProductUsageRecordsServer interface {
	Send(*InvalidProductUsageRecord) error
	grpc.ServerStream
}

type usageRecordServiceSearchInvalidProductUsageRecordsServer struct {
	grpc.ServerStream
}

func (x *usageRecordServiceSearchInvalidProductUsageRecordsServer) Send(m *InvalidProductUsageRecord) error {
	return x.ServerStream.SendMsg(m)
}

func _UsageRecordService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UsageRecordServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.UsageRecordService/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UsageRecordServiceServer).Ping(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// UsageRecordService_ServiceDesc is the grpc.ServiceDesc for UsageRecordService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UsageRecordService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.UsageRecordService",
	HandlerType: (*UsageRecordServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateProductUsageRecord",
			Handler:    _UsageRecordService_CreateProductUsageRecord_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _UsageRecordService_Ping_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SearchProductUsageRecords",
			Handler:       _UsageRecordService_SearchProductUsageRecords_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SearchInvalidProductUsageRecords",
			Handler:       _UsageRecordService_SearchInvalidProductUsageRecords_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "usage.proto",
}
