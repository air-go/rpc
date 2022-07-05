// Code generated by MockGen. DO NOT EDIT.
// Source: ./listener/listener.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	agollo "github.com/apolloconfig/agollo/v4"
	storage "github.com/apolloconfig/agollo/v4/storage"
	gomock "github.com/golang/mock/gomock"
)

// MockListener is a mock of Listener interface.
type MockListener struct {
	ctrl     *gomock.Controller
	recorder *MockListenerMockRecorder
}

// MockListenerMockRecorder is the mock recorder for MockListener.
type MockListenerMockRecorder struct {
	mock *MockListener
}

// NewMockListener creates a new mock instance.
func NewMockListener(ctrl *gomock.Controller) *MockListener {
	mock := &MockListener{ctrl: ctrl}
	mock.recorder = &MockListenerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockListener) EXPECT() *MockListenerMockRecorder {
	return m.recorder
}

// InitConfig mocks base method.
func (m *MockListener) InitConfig(client agollo.Client) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "InitConfig", client)
}

// InitConfig indicates an expected call of InitConfig.
func (mr *MockListenerMockRecorder) InitConfig(client interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitConfig", reflect.TypeOf((*MockListener)(nil).InitConfig), client)
}

// OnChange mocks base method.
func (m *MockListener) OnChange(event *storage.ChangeEvent) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnChange", event)
}

// OnChange indicates an expected call of OnChange.
func (mr *MockListenerMockRecorder) OnChange(event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnChange", reflect.TypeOf((*MockListener)(nil).OnChange), event)
}

// OnNewestChange mocks base method.
func (m *MockListener) OnNewestChange(event *storage.FullChangeEvent) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "OnNewestChange", event)
}

// OnNewestChange indicates an expected call of OnNewestChange.
func (mr *MockListenerMockRecorder) OnNewestChange(event interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OnNewestChange", reflect.TypeOf((*MockListener)(nil).OnNewestChange), event)
}