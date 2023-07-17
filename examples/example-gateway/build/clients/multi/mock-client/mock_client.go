// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/uber/zanzibar/v2/examples/example-gateway/build/clients/multi (interfaces: Client)

// Package clientmock is a generated GoMock package.
package clientmock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	zanzibar "github.com/uber/zanzibar/v2/runtime"
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

// HTTPClient mocks base method.
func (m *MockClient) HTTPClient() *zanzibar.HTTPClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HTTPClient")
	ret0, _ := ret[0].(*zanzibar.HTTPClient)
	return ret0
}

// HTTPClient indicates an expected call of HTTPClient.
func (mr *MockClientMockRecorder) HTTPClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HTTPClient", reflect.TypeOf((*MockClient)(nil).HTTPClient))
}

// HelloA mocks base method.
func (m *MockClient) HelloA(arg0 context.Context, arg1 map[string]string) (context.Context, string, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HelloA", arg0, arg1)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// HelloA indicates an expected call of HelloA.
func (mr *MockClientMockRecorder) HelloA(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HelloA", reflect.TypeOf((*MockClient)(nil).HelloA), arg0, arg1)
}

// HelloB mocks base method.
func (m *MockClient) HelloB(arg0 context.Context, arg1 map[string]string) (context.Context, string, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HelloB", arg0, arg1)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// HelloB indicates an expected call of HelloB.
func (mr *MockClientMockRecorder) HelloB(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HelloB", reflect.TypeOf((*MockClient)(nil).HelloB), arg0, arg1)
}
