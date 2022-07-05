// Code generated by MockGen. DO NOT EDIT.
// Source: /home/users/weihaoyu/go/pkg/mod/github.com/apolloconfig/agollo/v4@v4.1.1/client.go

// Package mock is a generated GoMock package.
package mock

import (
	list "container/list"
	reflect "reflect"

	agcache "github.com/apolloconfig/agollo/v4/agcache"
	storage "github.com/apolloconfig/agollo/v4/storage"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// AddChangeListener mocks base method.
func (m *MockClient) AddChangeListener(listener storage.ChangeListener) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AddChangeListener", listener)
}

// AddChangeListener indicates an expected call of AddChangeListener.
func (mr *MockClientMockRecorder) AddChangeListener(listener interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddChangeListener", reflect.TypeOf((*MockClient)(nil).AddChangeListener), listener)
}

// GetApolloConfigCache mocks base method.
func (m *MockClient) GetApolloConfigCache() agcache.CacheInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetApolloConfigCache")
	ret0, _ := ret[0].(agcache.CacheInterface)
	return ret0
}

// GetApolloConfigCache indicates an expected call of GetApolloConfigCache.
func (mr *MockClientMockRecorder) GetApolloConfigCache() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetApolloConfigCache", reflect.TypeOf((*MockClient)(nil).GetApolloConfigCache))
}

// GetBoolValue mocks base method.
func (m *MockClient) GetBoolValue(key string, defaultValue bool) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBoolValue", key, defaultValue)
	ret0, _ := ret[0].(bool)
	return ret0
}

// GetBoolValue indicates an expected call of GetBoolValue.
func (mr *MockClientMockRecorder) GetBoolValue(key, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBoolValue", reflect.TypeOf((*MockClient)(nil).GetBoolValue), key, defaultValue)
}

// GetChangeListeners mocks base method.
func (m *MockClient) GetChangeListeners() *list.List {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetChangeListeners")
	ret0, _ := ret[0].(*list.List)
	return ret0
}

// GetChangeListeners indicates an expected call of GetChangeListeners.
func (mr *MockClientMockRecorder) GetChangeListeners() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetChangeListeners", reflect.TypeOf((*MockClient)(nil).GetChangeListeners))
}

// GetConfig mocks base method.
func (m *MockClient) GetConfig(namespace string) *storage.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfig", namespace)
	ret0, _ := ret[0].(*storage.Config)
	return ret0
}

// GetConfig indicates an expected call of GetConfig.
func (mr *MockClientMockRecorder) GetConfig(namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfig", reflect.TypeOf((*MockClient)(nil).GetConfig), namespace)
}

// GetConfigAndInit mocks base method.
func (m *MockClient) GetConfigAndInit(namespace string) *storage.Config {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfigAndInit", namespace)
	ret0, _ := ret[0].(*storage.Config)
	return ret0
}

// GetConfigAndInit indicates an expected call of GetConfigAndInit.
func (mr *MockClientMockRecorder) GetConfigAndInit(namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfigAndInit", reflect.TypeOf((*MockClient)(nil).GetConfigAndInit), namespace)
}

// GetConfigCache mocks base method.
func (m *MockClient) GetConfigCache(namespace string) agcache.CacheInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfigCache", namespace)
	ret0, _ := ret[0].(agcache.CacheInterface)
	return ret0
}

// GetConfigCache indicates an expected call of GetConfigCache.
func (mr *MockClientMockRecorder) GetConfigCache(namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfigCache", reflect.TypeOf((*MockClient)(nil).GetConfigCache), namespace)
}

// GetDefaultConfigCache mocks base method.
func (m *MockClient) GetDefaultConfigCache() agcache.CacheInterface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDefaultConfigCache")
	ret0, _ := ret[0].(agcache.CacheInterface)
	return ret0
}

// GetDefaultConfigCache indicates an expected call of GetDefaultConfigCache.
func (mr *MockClientMockRecorder) GetDefaultConfigCache() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDefaultConfigCache", reflect.TypeOf((*MockClient)(nil).GetDefaultConfigCache))
}

// GetFloatValue mocks base method.
func (m *MockClient) GetFloatValue(key string, defaultValue float64) float64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFloatValue", key, defaultValue)
	ret0, _ := ret[0].(float64)
	return ret0
}

// GetFloatValue indicates an expected call of GetFloatValue.
func (mr *MockClientMockRecorder) GetFloatValue(key, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFloatValue", reflect.TypeOf((*MockClient)(nil).GetFloatValue), key, defaultValue)
}

// GetIntSliceValue mocks base method.
func (m *MockClient) GetIntSliceValue(key string, defaultValue []int) []int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIntSliceValue", key, defaultValue)
	ret0, _ := ret[0].([]int)
	return ret0
}

// GetIntSliceValue indicates an expected call of GetIntSliceValue.
func (mr *MockClientMockRecorder) GetIntSliceValue(key, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIntSliceValue", reflect.TypeOf((*MockClient)(nil).GetIntSliceValue), key, defaultValue)
}

// GetIntValue mocks base method.
func (m *MockClient) GetIntValue(key string, defaultValue int) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetIntValue", key, defaultValue)
	ret0, _ := ret[0].(int)
	return ret0
}

// GetIntValue indicates an expected call of GetIntValue.
func (mr *MockClientMockRecorder) GetIntValue(key, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetIntValue", reflect.TypeOf((*MockClient)(nil).GetIntValue), key, defaultValue)
}

// GetStringSliceValue mocks base method.
func (m *MockClient) GetStringSliceValue(key string, defaultValue []string) []string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStringSliceValue", key, defaultValue)
	ret0, _ := ret[0].([]string)
	return ret0
}

// GetStringSliceValue indicates an expected call of GetStringSliceValue.
func (mr *MockClientMockRecorder) GetStringSliceValue(key, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStringSliceValue", reflect.TypeOf((*MockClient)(nil).GetStringSliceValue), key, defaultValue)
}

// GetStringValue mocks base method.
func (m *MockClient) GetStringValue(key, defaultValue string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStringValue", key, defaultValue)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetStringValue indicates an expected call of GetStringValue.
func (mr *MockClientMockRecorder) GetStringValue(key, defaultValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStringValue", reflect.TypeOf((*MockClient)(nil).GetStringValue), key, defaultValue)
}

// GetValue mocks base method.
func (m *MockClient) GetValue(key string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValue", key)
	ret0, _ := ret[0].(string)
	return ret0
}

// GetValue indicates an expected call of GetValue.
func (mr *MockClientMockRecorder) GetValue(key interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValue", reflect.TypeOf((*MockClient)(nil).GetValue), key)
}

// RemoveChangeListener mocks base method.
func (m *MockClient) RemoveChangeListener(listener storage.ChangeListener) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RemoveChangeListener", listener)
}

// RemoveChangeListener indicates an expected call of RemoveChangeListener.
func (mr *MockClientMockRecorder) RemoveChangeListener(listener interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveChangeListener", reflect.TypeOf((*MockClient)(nil).RemoveChangeListener), listener)
}

// UseEventDispatch mocks base method.
func (m *MockClient) UseEventDispatch() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UseEventDispatch")
}

// UseEventDispatch indicates an expected call of UseEventDispatch.
func (mr *MockClientMockRecorder) UseEventDispatch() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UseEventDispatch", reflect.TypeOf((*MockClient)(nil).UseEventDispatch))
}
