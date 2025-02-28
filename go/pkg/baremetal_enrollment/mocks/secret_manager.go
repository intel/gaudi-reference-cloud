// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets (interfaces: SecretManager)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	api "github.com/hashicorp/vault/api"
)

// MockSecretManager is a mock of SecretManager interface.
type MockSecretManager struct {
	ctrl     *gomock.Controller
	recorder *MockSecretManagerMockRecorder
}

// MockSecretManagerMockRecorder is the mock recorder for MockSecretManager.
type MockSecretManagerMockRecorder struct {
	mock *MockSecretManager
}

// NewMockSecretManager creates a new mock instance.
func NewMockSecretManager(ctrl *gomock.Controller) *MockSecretManager {
	mock := &MockSecretManager{ctrl: ctrl}
	mock.recorder = &MockSecretManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSecretManager) EXPECT() *MockSecretManagerMockRecorder {
	return m.recorder
}

// DeleteBMCSecrets mocks base method.
func (m *MockSecretManager) DeleteBMCSecrets(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteBMCSecrets", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteBMCSecrets indicates an expected call of DeleteBMCSecrets.
func (mr *MockSecretManagerMockRecorder) DeleteBMCSecrets(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteBMCSecrets", reflect.TypeOf((*MockSecretManager)(nil).DeleteBMCSecrets), arg0, arg1)
}

// GetBMCBIOSPassword mocks base method.
func (m *MockSecretManager) GetBMCBIOSPassword(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBMCBIOSPassword", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBMCBIOSPassword indicates an expected call of GetBMCBIOSPassword.
func (mr *MockSecretManagerMockRecorder) GetBMCBIOSPassword(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBMCBIOSPassword", reflect.TypeOf((*MockSecretManager)(nil).GetBMCBIOSPassword), arg0, arg1)
}

// GetBMCCredentials mocks base method.
func (m *MockSecretManager) GetBMCCredentials(arg0 context.Context, arg1 string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBMCCredentials", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetBMCCredentials indicates an expected call of GetBMCCredentials.
func (mr *MockSecretManagerMockRecorder) GetBMCCredentials(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBMCCredentials", reflect.TypeOf((*MockSecretManager)(nil).GetBMCCredentials), arg0, arg1)
}

// GetControlPlaneSecrets mocks base method.
func (m *MockSecretManager) GetControlPlaneSecrets(arg0 context.Context, arg1 string) (*api.KVSecret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetControlPlaneSecrets", arg0, arg1)
	ret0, _ := ret[0].(*api.KVSecret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetControlPlaneSecrets indicates an expected call of GetControlPlaneSecrets.
func (mr *MockSecretManagerMockRecorder) GetControlPlaneSecrets(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetControlPlaneSecrets", reflect.TypeOf((*MockSecretManager)(nil).GetControlPlaneSecrets), arg0, arg1)
}

// GetDDICredentials mocks base method.
func (m *MockSecretManager) GetDDICredentials(arg0 context.Context, arg1 string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDDICredentials", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetDDICredentials indicates an expected call of GetDDICredentials.
func (mr *MockSecretManagerMockRecorder) GetDDICredentials(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDDICredentials", reflect.TypeOf((*MockSecretManager)(nil).GetDDICredentials), arg0, arg1)
}

// GetEnrollBasicAuth mocks base method.
func (m *MockSecretManager) GetEnrollBasicAuth(arg0 context.Context, arg1 string, arg2 bool) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEnrollBasicAuth", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetEnrollBasicAuth indicates an expected call of GetEnrollBasicAuth.
func (mr *MockSecretManagerMockRecorder) GetEnrollBasicAuth(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEnrollBasicAuth", reflect.TypeOf((*MockSecretManager)(nil).GetEnrollBasicAuth), arg0, arg1, arg2)
}

// GetIPAImageSSHPrivateKey mocks base method.
func (m *MockSecretManager) GetIPAImageSSHPrivateKey(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIPAImageSSHPrivateKey", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetIPAImageSSHPrivateKey indicates an expected call of GetIPAImageSSHPrivateKey.
func (mr *MockSecretManagerMockRecorder) GetIPAImageSSHPrivateKey(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIPAImageSSHPrivateKey", reflect.TypeOf((*MockSecretManager)(nil).GetIPAImageSSHPrivateKey), arg0, arg1)
}

// GetNetBoxAPIToken mocks base method.
func (m *MockSecretManager) GetNetBoxAPIToken(arg0 context.Context, arg1 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNetBoxAPIToken", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNetBoxAPIToken indicates an expected call of GetNetBoxAPIToken.
func (mr *MockSecretManagerMockRecorder) GetNetBoxAPIToken(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNetBoxAPIToken", reflect.TypeOf((*MockSecretManager)(nil).GetNetBoxAPIToken), arg0, arg1)
}

// PutBMCSecrets mocks base method.
func (m *MockSecretManager) PutBMCSecrets(arg0 context.Context, arg1 string, arg2 map[string]interface{}) (*api.KVSecret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutBMCSecrets", arg0, arg1, arg2)
	ret0, _ := ret[0].(*api.KVSecret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PutBMCSecrets indicates an expected call of PutBMCSecrets.
func (mr *MockSecretManagerMockRecorder) PutBMCSecrets(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutBMCSecrets", reflect.TypeOf((*MockSecretManager)(nil).PutBMCSecrets), arg0, arg1, arg2)
}

// ValidateVaultClient mocks base method.
func (m *MockSecretManager) ValidateVaultClient(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateVaultClient", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateVaultClient indicates an expected call of ValidateVaultClient.
func (mr *MockSecretManagerMockRecorder) ValidateVaultClient(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateVaultClient", reflect.TypeOf((*MockSecretManager)(nil).ValidateVaultClient), arg0)
}
