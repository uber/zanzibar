// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/uber/zanzibar/examples/selective-gateway/build/clients/echo (interfaces: Client)

// Package clientmock is a generated GoMock package.
package clientmock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	echo "github.com/uber/zanzibar/examples/selective-gateway/build/proto-gen/clients/echo"
	yarpc "go.uber.org/yarpc"
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

// EchoEcho mocks base method.
func (m *MockClient) EchoEcho(arg0 context.Context, arg1 *echo.Request, arg2 ...yarpc.CallOption) (*echo.Response, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "EchoEcho", varargs...)
	ret0, _ := ret[0].(*echo.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EchoEcho indicates an expected call of EchoEcho.
func (mr *MockClientMockRecorder) EchoEcho(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoEcho", reflect.TypeOf((*MockClient)(nil).EchoEcho), varargs...)
}
