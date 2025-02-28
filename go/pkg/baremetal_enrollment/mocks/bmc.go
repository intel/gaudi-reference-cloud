// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc (interfaces: Interface)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	bmc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/bmc"
	mygofish "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/mygofish"
	redfish "github.com/stmcginnis/gofish/redfish"
)

// MockBMCInterface is a mock of Interface interface.
type MockBMCInterface struct {
	ctrl     *gomock.Controller
	recorder *MockBMCInterfaceMockRecorder
}

// MockBMCInterfaceMockRecorder is the mock recorder for MockBMCInterface.
type MockBMCInterfaceMockRecorder struct {
	mock *MockBMCInterface
}

// NewMockBMCInterface creates a new mock instance.
func NewMockBMCInterface(ctrl *gomock.Controller) *MockBMCInterface {
	mock := &MockBMCInterface{ctrl: ctrl}
	mock.recorder = &MockBMCInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBMCInterface) EXPECT() *MockBMCInterfaceMockRecorder {
	return m.recorder
}

// ConfigureNTP mocks base method.
func (m *MockBMCInterface) ConfigureNTP(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConfigureNTP", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ConfigureNTP indicates an expected call of ConfigureNTP.
func (mr *MockBMCInterfaceMockRecorder) ConfigureNTP(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConfigureNTP", reflect.TypeOf((*MockBMCInterface)(nil).ConfigureNTP), arg0)
}

// CreateAccount mocks base method.
func (m *MockBMCInterface) CreateAccount(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAccount", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateAccount indicates an expected call of CreateAccount.
func (mr *MockBMCInterfaceMockRecorder) CreateAccount(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAccount", reflect.TypeOf((*MockBMCInterface)(nil).CreateAccount), arg0, arg1, arg2)
}

// DisableHCI mocks base method.
func (m *MockBMCInterface) DisableHCI(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DisableHCI", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DisableHCI indicates an expected call of DisableHCI.
func (mr *MockBMCInterfaceMockRecorder) DisableHCI(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DisableHCI", reflect.TypeOf((*MockBMCInterface)(nil).DisableHCI), arg0)
}

// DisableKCS mocks base method.
func (m *MockBMCInterface) DisableKCS(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DisableKCS", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DisableKCS indicates an expected call of DisableKCS.
func (mr *MockBMCInterfaceMockRecorder) DisableKCS(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DisableKCS", reflect.TypeOf((*MockBMCInterface)(nil).DisableKCS), arg0)
}

// EnableHCI mocks base method.
func (m *MockBMCInterface) EnableHCI(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnableHCI", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnableHCI indicates an expected call of EnableHCI.
func (mr *MockBMCInterfaceMockRecorder) EnableHCI(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnableHCI", reflect.TypeOf((*MockBMCInterface)(nil).EnableHCI), arg0)
}

// EnableKCS mocks base method.
func (m *MockBMCInterface) EnableKCS(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnableKCS", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnableKCS indicates an expected call of EnableKCS.
func (mr *MockBMCInterfaceMockRecorder) EnableKCS(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnableKCS", reflect.TypeOf((*MockBMCInterface)(nil).EnableKCS), arg0)
}

// GPUDiscovery mocks base method.
func (m *MockBMCInterface) GPUDiscovery(arg0 context.Context) (int, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GPUDiscovery", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GPUDiscovery indicates an expected call of GPUDiscovery.
func (mr *MockBMCInterfaceMockRecorder) GPUDiscovery(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GPUDiscovery", reflect.TypeOf((*MockBMCInterface)(nil).GPUDiscovery), arg0)
}

// GetBMCPowerState mocks base method.
func (m *MockBMCInterface) GetBMCPowerState(arg0 context.Context) (redfish.PowerState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBMCPowerState", arg0)
	ret0, _ := ret[0].(redfish.PowerState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBMCPowerState indicates an expected call of GetBMCPowerState.
func (mr *MockBMCInterfaceMockRecorder) GetBMCPowerState(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBMCPowerState", reflect.TypeOf((*MockBMCInterface)(nil).GetBMCPowerState), arg0)
}

// GetClient mocks base method.
func (m *MockBMCInterface) GetClient() mygofish.GoFishClientAccessor {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClient")
	ret0, _ := ret[0].(mygofish.GoFishClientAccessor)
	return ret0
}

// GetClient indicates an expected call of GetClient.
func (mr *MockBMCInterfaceMockRecorder) GetClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClient", reflect.TypeOf((*MockBMCInterface)(nil).GetClient))
}

// GetHostBMCAddress mocks base method.
func (m *MockBMCInterface) GetHostBMCAddress() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHostBMCAddress")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHostBMCAddress indicates an expected call of GetHostBMCAddress.
func (mr *MockBMCInterfaceMockRecorder) GetHostBMCAddress() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHostBMCAddress", reflect.TypeOf((*MockBMCInterface)(nil).GetHostBMCAddress))
}

// GetHostCPU mocks base method.
func (m *MockBMCInterface) GetHostCPU(arg0 context.Context) (*bmc.CPUInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHostCPU", arg0)
	ret0, _ := ret[0].(*bmc.CPUInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHostCPU indicates an expected call of GetHostCPU.
func (mr *MockBMCInterfaceMockRecorder) GetHostCPU(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHostCPU", reflect.TypeOf((*MockBMCInterface)(nil).GetHostCPU), arg0)
}

// GetHostMACAddress mocks base method.
func (m *MockBMCInterface) GetHostMACAddress(arg0 context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHostMACAddress", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHostMACAddress indicates an expected call of GetHostMACAddress.
func (mr *MockBMCInterfaceMockRecorder) GetHostMACAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHostMACAddress", reflect.TypeOf((*MockBMCInterface)(nil).GetHostMACAddress), arg0)
}

// GetHwType mocks base method.
func (m *MockBMCInterface) GetHwType() bmc.HWType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHwType")
	ret0, _ := ret[0].(bmc.HWType)
	return ret0
}

// GetHwType indicates an expected call of GetHwType.
func (mr *MockBMCInterfaceMockRecorder) GetHwType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHwType", reflect.TypeOf((*MockBMCInterface)(nil).GetHwType))
}

// HBMDiscovery mocks base method.
func (m *MockBMCInterface) HBMDiscovery(arg0 context.Context) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HBMDiscovery", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HBMDiscovery indicates an expected call of HBMDiscovery.
func (mr *MockBMCInterfaceMockRecorder) HBMDiscovery(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HBMDiscovery", reflect.TypeOf((*MockBMCInterface)(nil).HBMDiscovery), arg0)
}

// IsIntelPlatform mocks base method.
func (m *MockBMCInterface) IsIntelPlatform() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsIntelPlatform")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsIntelPlatform indicates an expected call of IsIntelPlatform.
func (mr *MockBMCInterfaceMockRecorder) IsIntelPlatform() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsIntelPlatform", reflect.TypeOf((*MockBMCInterface)(nil).IsIntelPlatform))
}

// IsVirtual mocks base method.
func (m *MockBMCInterface) IsVirtual() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsVirtual")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsVirtual indicates an expected call of IsVirtual.
func (mr *MockBMCInterfaceMockRecorder) IsVirtual() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsVirtual", reflect.TypeOf((*MockBMCInterface)(nil).IsVirtual))
}

// PowerOffBMC mocks base method.
func (m *MockBMCInterface) PowerOffBMC(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PowerOffBMC", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// PowerOffBMC indicates an expected call of PowerOffBMC.
func (mr *MockBMCInterfaceMockRecorder) PowerOffBMC(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PowerOffBMC", reflect.TypeOf((*MockBMCInterface)(nil).PowerOffBMC), arg0)
}

// PowerOnBMC mocks base method.
func (m *MockBMCInterface) PowerOnBMC(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PowerOnBMC", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// PowerOnBMC indicates an expected call of PowerOnBMC.
func (mr *MockBMCInterfaceMockRecorder) PowerOnBMC(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PowerOnBMC", reflect.TypeOf((*MockBMCInterface)(nil).PowerOnBMC), arg0)
}

// SanitizeBMCBootOrder mocks base method.
func (m *MockBMCInterface) SanitizeBMCBootOrder(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SanitizeBMCBootOrder", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SanitizeBMCBootOrder indicates an expected call of SanitizeBMCBootOrder.
func (mr *MockBMCInterfaceMockRecorder) SanitizeBMCBootOrder(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SanitizeBMCBootOrder", reflect.TypeOf((*MockBMCInterface)(nil).SanitizeBMCBootOrder), arg0)
}

// SetFanSpeed mocks base method.
func (m *MockBMCInterface) SetFanSpeed(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetFanSpeed", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetFanSpeed indicates an expected call of SetFanSpeed.
func (mr *MockBMCInterfaceMockRecorder) SetFanSpeed(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFanSpeed", reflect.TypeOf((*MockBMCInterface)(nil).SetFanSpeed), arg0)
}

// UpdateAccount mocks base method.
func (m *MockBMCInterface) UpdateAccount(arg0 context.Context, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAccount", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAccount indicates an expected call of UpdateAccount.
func (mr *MockBMCInterfaceMockRecorder) UpdateAccount(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAccount", reflect.TypeOf((*MockBMCInterface)(nil).UpdateAccount), arg0, arg1, arg2)
}

// VerifyPlatformFirmwareResilience mocks base method.
func (m *MockBMCInterface) VerifyPlatformFirmwareResilience(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyPlatformFirmwareResilience", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyPlatformFirmwareResilience indicates an expected call of VerifyPlatformFirmwareResilience.
func (mr *MockBMCInterfaceMockRecorder) VerifyPlatformFirmwareResilience(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyPlatformFirmwareResilience", reflect.TypeOf((*MockBMCInterface)(nil).VerifyPlatformFirmwareResilience), arg0)
}
