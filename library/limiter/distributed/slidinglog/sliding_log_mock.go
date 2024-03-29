// Code generated by MockGen. DO NOT EDIT.
// Source: sliding_log.go

// Package slidinglog is a generated GoMock package.
package slidinglog

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockSlidingLog is a mock of SlidingLog interface.
type MockSlidingLog struct {
	ctrl     *gomock.Controller
	recorder *MockSlidingLogMockRecorder
}

// MockSlidingLogMockRecorder is the mock recorder for MockSlidingLog.
type MockSlidingLogMockRecorder struct {
	mock *MockSlidingLog
}

// NewMockSlidingLog creates a new mock instance.
func NewMockSlidingLog(ctrl *gomock.Controller) *MockSlidingLog {
	mock := &MockSlidingLog{ctrl: ctrl}
	mock.recorder = &MockSlidingLogMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSlidingLog) EXPECT() *MockSlidingLogMockRecorder {
	return m.recorder
}

// Allow mocks base method.
func (m *MockSlidingLog) Allow(ctx context.Context, key string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Allow", ctx, key)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Allow indicates an expected call of Allow.
func (mr *MockSlidingLogMockRecorder) Allow(ctx, key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Allow", reflect.TypeOf((*MockSlidingLog)(nil).Allow), ctx, key)
}

// SetLimit mocks base method.
func (m *MockSlidingLog) SetLimit(limit int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetLimit", limit)
}

// SetLimit indicates an expected call of SetLimit.
func (mr *MockSlidingLogMockRecorder) SetLimit(limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLimit", reflect.TypeOf((*MockSlidingLog)(nil).SetLimit), limit)
}

// SetWindow mocks base method.
func (m *MockSlidingLog) SetWindow(w time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetWindow", w)
}

// SetWindow indicates an expected call of SetWindow.
func (mr *MockSlidingLogMockRecorder) SetWindow(w interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetWindow", reflect.TypeOf((*MockSlidingLog)(nil).SetWindow), w)
}
