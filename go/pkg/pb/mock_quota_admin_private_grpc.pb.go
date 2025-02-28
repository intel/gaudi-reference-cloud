// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/pb/quota_admin_private_grpc.pb.go

// Package pb is a generated GoMock package.
package pb

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// MockQuotaManagementPrivateServiceClient is a mock of QuotaManagementPrivateServiceClient interface.
type MockQuotaManagementPrivateServiceClient struct {
	ctrl     *gomock.Controller
	recorder *MockQuotaManagementPrivateServiceClientMockRecorder
}

// MockQuotaManagementPrivateServiceClientMockRecorder is the mock recorder for MockQuotaManagementPrivateServiceClient.
type MockQuotaManagementPrivateServiceClientMockRecorder struct {
	mock *MockQuotaManagementPrivateServiceClient
}

// NewMockQuotaManagementPrivateServiceClient creates a new mock instance.
func NewMockQuotaManagementPrivateServiceClient(ctrl *gomock.Controller) *MockQuotaManagementPrivateServiceClient {
	mock := &MockQuotaManagementPrivateServiceClient{ctrl: ctrl}
	mock.recorder = &MockQuotaManagementPrivateServiceClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQuotaManagementPrivateServiceClient) EXPECT() *MockQuotaManagementPrivateServiceClientMockRecorder {
	return m.recorder
}

// GetResourceQuotaPrivate mocks base method.
func (m *MockQuotaManagementPrivateServiceClient) GetResourceQuotaPrivate(ctx context.Context, in *ServiceQuotaResourceRequestPrivate, opts ...grpc.CallOption) (*ServiceQuotasPrivate, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetResourceQuotaPrivate", varargs...)
	ret0, _ := ret[0].(*ServiceQuotasPrivate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetResourceQuotaPrivate indicates an expected call of GetResourceQuotaPrivate.
func (mr *MockQuotaManagementPrivateServiceClientMockRecorder) GetResourceQuotaPrivate(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResourceQuotaPrivate", reflect.TypeOf((*MockQuotaManagementPrivateServiceClient)(nil).GetResourceQuotaPrivate), varargs...)
}

// PingPrivate mocks base method.
func (m *MockQuotaManagementPrivateServiceClient) PingPrivate(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "PingPrivate", varargs...)
	ret0, _ := ret[0].(*emptypb.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PingPrivate indicates an expected call of PingPrivate.
func (mr *MockQuotaManagementPrivateServiceClientMockRecorder) PingPrivate(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingPrivate", reflect.TypeOf((*MockQuotaManagementPrivateServiceClient)(nil).PingPrivate), varargs...)
}

// MockQuotaManagementPrivateServiceServer is a mock of QuotaManagementPrivateServiceServer interface.
type MockQuotaManagementPrivateServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockQuotaManagementPrivateServiceServerMockRecorder
}

// MockQuotaManagementPrivateServiceServerMockRecorder is the mock recorder for MockQuotaManagementPrivateServiceServer.
type MockQuotaManagementPrivateServiceServerMockRecorder struct {
	mock *MockQuotaManagementPrivateServiceServer
}

// NewMockQuotaManagementPrivateServiceServer creates a new mock instance.
func NewMockQuotaManagementPrivateServiceServer(ctrl *gomock.Controller) *MockQuotaManagementPrivateServiceServer {
	mock := &MockQuotaManagementPrivateServiceServer{ctrl: ctrl}
	mock.recorder = &MockQuotaManagementPrivateServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockQuotaManagementPrivateServiceServer) EXPECT() *MockQuotaManagementPrivateServiceServerMockRecorder {
	return m.recorder
}

// GetResourceQuotaPrivate mocks base method.
func (m *MockQuotaManagementPrivateServiceServer) GetResourceQuotaPrivate(arg0 context.Context, arg1 *ServiceQuotaResourceRequestPrivate) (*ServiceQuotasPrivate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResourceQuotaPrivate", arg0, arg1)
	ret0, _ := ret[0].(*ServiceQuotasPrivate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetResourceQuotaPrivate indicates an expected call of GetResourceQuotaPrivate.
func (mr *MockQuotaManagementPrivateServiceServerMockRecorder) GetResourceQuotaPrivate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResourceQuotaPrivate", reflect.TypeOf((*MockQuotaManagementPrivateServiceServer)(nil).GetResourceQuotaPrivate), arg0, arg1)
}

// PingPrivate mocks base method.
func (m *MockQuotaManagementPrivateServiceServer) PingPrivate(arg0 context.Context, arg1 *emptypb.Empty) (*emptypb.Empty, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingPrivate", arg0, arg1)
	ret0, _ := ret[0].(*emptypb.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PingPrivate indicates an expected call of PingPrivate.
func (mr *MockQuotaManagementPrivateServiceServerMockRecorder) PingPrivate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingPrivate", reflect.TypeOf((*MockQuotaManagementPrivateServiceServer)(nil).PingPrivate), arg0, arg1)
}

// mustEmbedUnimplementedQuotaManagementPrivateServiceServer mocks base method.
func (m *MockQuotaManagementPrivateServiceServer) mustEmbedUnimplementedQuotaManagementPrivateServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedQuotaManagementPrivateServiceServer")
}

// mustEmbedUnimplementedQuotaManagementPrivateServiceServer indicates an expected call of mustEmbedUnimplementedQuotaManagementPrivateServiceServer.
func (mr *MockQuotaManagementPrivateServiceServerMockRecorder) mustEmbedUnimplementedQuotaManagementPrivateServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedQuotaManagementPrivateServiceServer", reflect.TypeOf((*MockQuotaManagementPrivateServiceServer)(nil).mustEmbedUnimplementedQuotaManagementPrivateServiceServer))
}

// MockUnsafeQuotaManagementPrivateServiceServer is a mock of UnsafeQuotaManagementPrivateServiceServer interface.
type MockUnsafeQuotaManagementPrivateServiceServer struct {
	ctrl     *gomock.Controller
	recorder *MockUnsafeQuotaManagementPrivateServiceServerMockRecorder
}

// MockUnsafeQuotaManagementPrivateServiceServerMockRecorder is the mock recorder for MockUnsafeQuotaManagementPrivateServiceServer.
type MockUnsafeQuotaManagementPrivateServiceServerMockRecorder struct {
	mock *MockUnsafeQuotaManagementPrivateServiceServer
}

// NewMockUnsafeQuotaManagementPrivateServiceServer creates a new mock instance.
func NewMockUnsafeQuotaManagementPrivateServiceServer(ctrl *gomock.Controller) *MockUnsafeQuotaManagementPrivateServiceServer {
	mock := &MockUnsafeQuotaManagementPrivateServiceServer{ctrl: ctrl}
	mock.recorder = &MockUnsafeQuotaManagementPrivateServiceServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnsafeQuotaManagementPrivateServiceServer) EXPECT() *MockUnsafeQuotaManagementPrivateServiceServerMockRecorder {
	return m.recorder
}

// mustEmbedUnimplementedQuotaManagementPrivateServiceServer mocks base method.
func (m *MockUnsafeQuotaManagementPrivateServiceServer) mustEmbedUnimplementedQuotaManagementPrivateServiceServer() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "mustEmbedUnimplementedQuotaManagementPrivateServiceServer")
}

// mustEmbedUnimplementedQuotaManagementPrivateServiceServer indicates an expected call of mustEmbedUnimplementedQuotaManagementPrivateServiceServer.
func (mr *MockUnsafeQuotaManagementPrivateServiceServerMockRecorder) mustEmbedUnimplementedQuotaManagementPrivateServiceServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "mustEmbedUnimplementedQuotaManagementPrivateServiceServer", reflect.TypeOf((*MockUnsafeQuotaManagementPrivateServiceServer)(nil).mustEmbedUnimplementedQuotaManagementPrivateServiceServer))
}
