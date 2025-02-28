// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/quick_connect/secrets (interfaces: VaultHelper)

// Package secrets is a generated GoMock package.
package secrets

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	api "github.com/hashicorp/vault/api"
)

// MockVaultHelper is a mock of VaultHelper interface.
type MockVaultHelper struct {
	ctrl     *gomock.Controller
	recorder *MockVaultHelperMockRecorder
}

// MockVaultHelperMockRecorder is the mock recorder for MockVaultHelper.
type MockVaultHelperMockRecorder struct {
	mock *MockVaultHelper
}

// NewMockVaultHelper creates a new mock instance.
func NewMockVaultHelper(ctrl *gomock.Controller) *MockVaultHelper {
	mock := &MockVaultHelper{ctrl: ctrl}
	mock.recorder = &MockVaultHelperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockVaultHelper) EXPECT() *MockVaultHelperMockRecorder {
	return m.recorder
}

// getVaultAuthInfo mocks base method.
func (m *MockVaultHelper) getVaultAuthInfo(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "getVaultAuthInfo", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// getVaultAuthInfo indicates an expected call of getVaultAuthInfo.
func (mr *MockVaultHelperMockRecorder) getVaultAuthInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "getVaultAuthInfo", reflect.TypeOf((*MockVaultHelper)(nil).getVaultAuthInfo), arg0)
}

// getVaultClient mocks base method.
func (m *MockVaultHelper) getVaultClient(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "getVaultClient", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// getVaultClient indicates an expected call of getVaultClient.
func (mr *MockVaultHelperMockRecorder) getVaultClient(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "getVaultClient", reflect.TypeOf((*MockVaultHelper)(nil).getVaultClient), arg0)
}

// issueVaultCertificate mocks base method.
func (m *MockVaultHelper) issueVaultCertificate(arg0 context.Context, arg1 map[string]interface{}, arg2 bool) (*api.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "issueVaultCertificate", arg0, arg1, arg2)
	ret0, _ := ret[0].(*api.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// issueVaultCertificate indicates an expected call of issueVaultCertificate.
func (mr *MockVaultHelperMockRecorder) issueVaultCertificate(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "issueVaultCertificate", reflect.TypeOf((*MockVaultHelper)(nil).issueVaultCertificate), arg0, arg1, arg2)
}
