// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/cosmos/ibc-go/v10/modules/core/05-port/types (interfaces: ICS4Wrapper)
//
// Generated by this command:
//
//	mockgen -package=mock -destination=./test/mock/ics4_wrapper.go github.com/cosmos/ibc-go/v10/modules/core/05-port/types ICS4Wrapper
//
// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	types "github.com/cosmos/cosmos-sdk/types"
	types0 "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	exported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	gomock "github.com/golang/mock/gomock"
)

// MockICS4Wrapper is a mock of ICS4Wrapper interface.
type MockICS4Wrapper struct {
	ctrl     *gomock.Controller
	recorder *MockICS4WrapperMockRecorder
}

// MockICS4WrapperMockRecorder is the mock recorder for MockICS4Wrapper.
type MockICS4WrapperMockRecorder struct {
	mock *MockICS4Wrapper
}

// NewMockICS4Wrapper creates a new mock instance.
func NewMockICS4Wrapper(ctrl *gomock.Controller) *MockICS4Wrapper {
	mock := &MockICS4Wrapper{ctrl: ctrl}
	mock.recorder = &MockICS4WrapperMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockICS4Wrapper) EXPECT() *MockICS4WrapperMockRecorder {
	return m.recorder
}

// GetAppVersion mocks base method.
func (m *MockICS4Wrapper) GetAppVersion(arg0 types.Context, arg1, arg2 string) (string, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAppVersion", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetAppVersion indicates an expected call of GetAppVersion.
func (mr *MockICS4WrapperMockRecorder) GetAppVersion(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAppVersion", reflect.TypeOf((*MockICS4Wrapper)(nil).GetAppVersion), arg0, arg1, arg2)
}

// SendPacket mocks base method.
func (m *MockICS4Wrapper) SendPacket(arg0 types.Context, arg1, arg2 string, arg3 types0.Height, arg4 uint64, arg5 []byte) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendPacket", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendPacket indicates an expected call of SendPacket.
func (mr *MockICS4WrapperMockRecorder) SendPacket(arg0, arg1, arg2, arg3, arg4, arg5 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendPacket", reflect.TypeOf((*MockICS4Wrapper)(nil).SendPacket), arg0, arg1, arg2, arg3, arg4, arg5)
}

// WriteAcknowledgement mocks base method.
func (m *MockICS4Wrapper) WriteAcknowledgement(arg0 types.Context, arg1 exported.PacketI, arg2 exported.Acknowledgement) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteAcknowledgement", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteAcknowledgement indicates an expected call of WriteAcknowledgement.
func (mr *MockICS4WrapperMockRecorder) WriteAcknowledgement(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteAcknowledgement", reflect.TypeOf((*MockICS4Wrapper)(nil).WriteAcknowledgement), arg0, arg1, arg2)
}
