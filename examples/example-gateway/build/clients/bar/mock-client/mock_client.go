// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/uber/zanzibar/examples/example-gateway/build/clients/bar (interfaces: Client)

// Package clientmock is a generated GoMock package.
package clientmock

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	bar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/bar/bar"
	runtime "github.com/uber/zanzibar/runtime"
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

// ArgNotStruct mocks base method.
func (m *MockClient) ArgNotStruct(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ArgNotStruct_Args) (context.Context, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgNotStruct", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(map[string]string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ArgNotStruct indicates an expected call of ArgNotStruct.
func (mr *MockClientMockRecorder) ArgNotStruct(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgNotStruct", reflect.TypeOf((*MockClient)(nil).ArgNotStruct), arg0, arg1, arg2)
}

// ArgWithHeaders mocks base method.
func (m *MockClient) ArgWithHeaders(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ArgWithHeaders_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgWithHeaders", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ArgWithHeaders indicates an expected call of ArgWithHeaders.
func (mr *MockClientMockRecorder) ArgWithHeaders(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgWithHeaders", reflect.TypeOf((*MockClient)(nil).ArgWithHeaders), arg0, arg1, arg2)
}

// ArgWithManyQueryParams mocks base method.
func (m *MockClient) ArgWithManyQueryParams(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ArgWithManyQueryParams_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgWithManyQueryParams", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ArgWithManyQueryParams indicates an expected call of ArgWithManyQueryParams.
func (mr *MockClientMockRecorder) ArgWithManyQueryParams(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgWithManyQueryParams", reflect.TypeOf((*MockClient)(nil).ArgWithManyQueryParams), arg0, arg1, arg2)
}

// ArgWithNearDupQueryParams mocks base method.
func (m *MockClient) ArgWithNearDupQueryParams(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ArgWithNearDupQueryParams_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgWithNearDupQueryParams", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ArgWithNearDupQueryParams indicates an expected call of ArgWithNearDupQueryParams.
func (mr *MockClientMockRecorder) ArgWithNearDupQueryParams(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgWithNearDupQueryParams", reflect.TypeOf((*MockClient)(nil).ArgWithNearDupQueryParams), arg0, arg1, arg2)
}

// ArgWithNestedQueryParams mocks base method.
func (m *MockClient) ArgWithNestedQueryParams(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ArgWithNestedQueryParams_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgWithNestedQueryParams", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ArgWithNestedQueryParams indicates an expected call of ArgWithNestedQueryParams.
func (mr *MockClientMockRecorder) ArgWithNestedQueryParams(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgWithNestedQueryParams", reflect.TypeOf((*MockClient)(nil).ArgWithNestedQueryParams), arg0, arg1, arg2)
}

// ArgWithParams mocks base method.
func (m *MockClient) ArgWithParams(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ArgWithParams_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgWithParams", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ArgWithParams indicates an expected call of ArgWithParams.
func (mr *MockClientMockRecorder) ArgWithParams(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgWithParams", reflect.TypeOf((*MockClient)(nil).ArgWithParams), arg0, arg1, arg2)
}

// ArgWithParamsAndDuplicateFields mocks base method.
func (m *MockClient) ArgWithParamsAndDuplicateFields(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ArgWithParamsAndDuplicateFields_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgWithParamsAndDuplicateFields", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ArgWithParamsAndDuplicateFields indicates an expected call of ArgWithParamsAndDuplicateFields.
func (mr *MockClientMockRecorder) ArgWithParamsAndDuplicateFields(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgWithParamsAndDuplicateFields", reflect.TypeOf((*MockClient)(nil).ArgWithParamsAndDuplicateFields), arg0, arg1, arg2)
}

// ArgWithQueryHeader mocks base method.
func (m *MockClient) ArgWithQueryHeader(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ArgWithQueryHeader_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgWithQueryHeader", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ArgWithQueryHeader indicates an expected call of ArgWithQueryHeader.
func (mr *MockClientMockRecorder) ArgWithQueryHeader(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgWithQueryHeader", reflect.TypeOf((*MockClient)(nil).ArgWithQueryHeader), arg0, arg1, arg2)
}

// ArgWithQueryParams mocks base method.
func (m *MockClient) ArgWithQueryParams(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ArgWithQueryParams_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ArgWithQueryParams", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ArgWithQueryParams indicates an expected call of ArgWithQueryParams.
func (mr *MockClientMockRecorder) ArgWithQueryParams(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ArgWithQueryParams", reflect.TypeOf((*MockClient)(nil).ArgWithQueryParams), arg0, arg1, arg2)
}

// DeleteFoo mocks base method.
func (m *MockClient) DeleteFoo(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_DeleteFoo_Args) (context.Context, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFoo", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(map[string]string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// DeleteFoo indicates an expected call of DeleteFoo.
func (mr *MockClientMockRecorder) DeleteFoo(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFoo", reflect.TypeOf((*MockClient)(nil).DeleteFoo), arg0, arg1, arg2)
}

// DeleteWithBody mocks base method.
func (m *MockClient) DeleteWithBody(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_DeleteWithBody_Args) (context.Context, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteWithBody", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(map[string]string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// DeleteWithBody indicates an expected call of DeleteWithBody.
func (mr *MockClientMockRecorder) DeleteWithBody(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteWithBody", reflect.TypeOf((*MockClient)(nil).DeleteWithBody), arg0, arg1, arg2)
}

// DeleteWithQueryParams mocks base method.
func (m *MockClient) DeleteWithQueryParams(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_DeleteWithQueryParams_Args) (context.Context, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteWithQueryParams", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(map[string]string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// DeleteWithQueryParams indicates an expected call of DeleteWithQueryParams.
func (mr *MockClientMockRecorder) DeleteWithQueryParams(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteWithQueryParams", reflect.TypeOf((*MockClient)(nil).DeleteWithQueryParams), arg0, arg1, arg2)
}

// EchoBinary mocks base method.
func (m *MockClient) EchoBinary(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoBinary_Args) (context.Context, []byte, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoBinary", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoBinary indicates an expected call of EchoBinary.
func (mr *MockClientMockRecorder) EchoBinary(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoBinary", reflect.TypeOf((*MockClient)(nil).EchoBinary), arg0, arg1, arg2)
}

// EchoBool mocks base method.
func (m *MockClient) EchoBool(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoBool_Args) (context.Context, bool, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoBool", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoBool indicates an expected call of EchoBool.
func (mr *MockClientMockRecorder) EchoBool(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoBool", reflect.TypeOf((*MockClient)(nil).EchoBool), arg0, arg1, arg2)
}

// EchoDouble mocks base method.
func (m *MockClient) EchoDouble(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoDouble_Args) (context.Context, float64, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoDouble", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(float64)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoDouble indicates an expected call of EchoDouble.
func (mr *MockClientMockRecorder) EchoDouble(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoDouble", reflect.TypeOf((*MockClient)(nil).EchoDouble), arg0, arg1, arg2)
}

// EchoEnum mocks base method.
func (m *MockClient) EchoEnum(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoEnum_Args) (context.Context, bar.Fruit, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoEnum", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(bar.Fruit)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoEnum indicates an expected call of EchoEnum.
func (mr *MockClientMockRecorder) EchoEnum(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoEnum", reflect.TypeOf((*MockClient)(nil).EchoEnum), arg0, arg1, arg2)
}

// EchoI16 mocks base method.
func (m *MockClient) EchoI16(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoI16_Args) (context.Context, int16, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoI16", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(int16)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoI16 indicates an expected call of EchoI16.
func (mr *MockClientMockRecorder) EchoI16(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoI16", reflect.TypeOf((*MockClient)(nil).EchoI16), arg0, arg1, arg2)
}

// EchoI32 mocks base method.
func (m *MockClient) EchoI32(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoI32_Args) (context.Context, int32, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoI32", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(int32)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoI32 indicates an expected call of EchoI32.
func (mr *MockClientMockRecorder) EchoI32(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoI32", reflect.TypeOf((*MockClient)(nil).EchoI32), arg0, arg1, arg2)
}

// EchoI32Map mocks base method.
func (m *MockClient) EchoI32Map(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoI32Map_Args) (context.Context, map[int32]*bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoI32Map", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(map[int32]*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoI32Map indicates an expected call of EchoI32Map.
func (mr *MockClientMockRecorder) EchoI32Map(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoI32Map", reflect.TypeOf((*MockClient)(nil).EchoI32Map), arg0, arg1, arg2)
}

// EchoI64 mocks base method.
func (m *MockClient) EchoI64(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoI64_Args) (context.Context, int64, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoI64", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoI64 indicates an expected call of EchoI64.
func (mr *MockClientMockRecorder) EchoI64(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoI64", reflect.TypeOf((*MockClient)(nil).EchoI64), arg0, arg1, arg2)
}

// EchoI8 mocks base method.
func (m *MockClient) EchoI8(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoI8_Args) (context.Context, int8, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoI8", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(int8)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoI8 indicates an expected call of EchoI8.
func (mr *MockClientMockRecorder) EchoI8(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoI8", reflect.TypeOf((*MockClient)(nil).EchoI8), arg0, arg1, arg2)
}

// EchoString mocks base method.
func (m *MockClient) EchoString(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoString_Args) (context.Context, string, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoString", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoString indicates an expected call of EchoString.
func (mr *MockClientMockRecorder) EchoString(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoString", reflect.TypeOf((*MockClient)(nil).EchoString), arg0, arg1, arg2)
}

// EchoStringList mocks base method.
func (m *MockClient) EchoStringList(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoStringList_Args) (context.Context, []string, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoStringList", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].([]string)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoStringList indicates an expected call of EchoStringList.
func (mr *MockClientMockRecorder) EchoStringList(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoStringList", reflect.TypeOf((*MockClient)(nil).EchoStringList), arg0, arg1, arg2)
}

// EchoStringMap mocks base method.
func (m *MockClient) EchoStringMap(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoStringMap_Args) (context.Context, map[string]*bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoStringMap", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(map[string]*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoStringMap indicates an expected call of EchoStringMap.
func (mr *MockClientMockRecorder) EchoStringMap(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoStringMap", reflect.TypeOf((*MockClient)(nil).EchoStringMap), arg0, arg1, arg2)
}

// EchoStringSet mocks base method.
func (m *MockClient) EchoStringSet(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoStringSet_Args) (context.Context, map[string]struct{}, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoStringSet", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(map[string]struct{})
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoStringSet indicates an expected call of EchoStringSet.
func (mr *MockClientMockRecorder) EchoStringSet(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoStringSet", reflect.TypeOf((*MockClient)(nil).EchoStringSet), arg0, arg1, arg2)
}

// EchoStructList mocks base method.
func (m *MockClient) EchoStructList(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoStructList_Args) (context.Context, []*bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoStructList", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].([]*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoStructList indicates an expected call of EchoStructList.
func (mr *MockClientMockRecorder) EchoStructList(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoStructList", reflect.TypeOf((*MockClient)(nil).EchoStructList), arg0, arg1, arg2)
}

// EchoStructSet mocks base method.
func (m *MockClient) EchoStructSet(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoStructSet_Args) (context.Context, []*bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoStructSet", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].([]*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoStructSet indicates an expected call of EchoStructSet.
func (mr *MockClientMockRecorder) EchoStructSet(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoStructSet", reflect.TypeOf((*MockClient)(nil).EchoStructSet), arg0, arg1, arg2)
}

// EchoTypedef mocks base method.
func (m *MockClient) EchoTypedef(arg0 context.Context, arg1 map[string]string, arg2 *bar.Echo_EchoTypedef_Args) (context.Context, bar.UUID, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EchoTypedef", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(bar.UUID)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// EchoTypedef indicates an expected call of EchoTypedef.
func (mr *MockClientMockRecorder) EchoTypedef(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EchoTypedef", reflect.TypeOf((*MockClient)(nil).EchoTypedef), arg0, arg1, arg2)
}

// HTTPClient mocks base method.
func (m *MockClient) HTTPClient() *runtime.HTTPClient {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HTTPClient")
	ret0, _ := ret[0].(*runtime.HTTPClient)
	return ret0
}

// HTTPClient indicates an expected call of HTTPClient.
func (mr *MockClientMockRecorder) HTTPClient() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HTTPClient", reflect.TypeOf((*MockClient)(nil).HTTPClient))
}

// Hello mocks base method.
func (m *MockClient) Hello(arg0 context.Context, arg1 map[string]string) (context.Context, string, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Hello", arg0, arg1)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// Hello indicates an expected call of Hello.
func (mr *MockClientMockRecorder) Hello(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Hello", reflect.TypeOf((*MockClient)(nil).Hello), arg0, arg1)
}

// ListAndEnum mocks base method.
func (m *MockClient) ListAndEnum(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_ListAndEnum_Args) (context.Context, string, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListAndEnum", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ListAndEnum indicates an expected call of ListAndEnum.
func (mr *MockClientMockRecorder) ListAndEnum(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListAndEnum", reflect.TypeOf((*MockClient)(nil).ListAndEnum), arg0, arg1, arg2)
}

// MissingArg mocks base method.
func (m *MockClient) MissingArg(arg0 context.Context, arg1 map[string]string) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MissingArg", arg0, arg1)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// MissingArg indicates an expected call of MissingArg.
func (mr *MockClientMockRecorder) MissingArg(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MissingArg", reflect.TypeOf((*MockClient)(nil).MissingArg), arg0, arg1)
}

// NoRequest mocks base method.
func (m *MockClient) NoRequest(arg0 context.Context, arg1 map[string]string) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NoRequest", arg0, arg1)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// NoRequest indicates an expected call of NoRequest.
func (mr *MockClientMockRecorder) NoRequest(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NoRequest", reflect.TypeOf((*MockClient)(nil).NoRequest), arg0, arg1)
}

// Normal mocks base method.
func (m *MockClient) Normal(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_Normal_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Normal", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// Normal indicates an expected call of Normal.
func (mr *MockClientMockRecorder) Normal(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Normal", reflect.TypeOf((*MockClient)(nil).Normal), arg0, arg1, arg2)
}

// NormalRecur mocks base method.
func (m *MockClient) NormalRecur(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_NormalRecur_Args) (context.Context, *bar.BarResponseRecur, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NormalRecur", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponseRecur)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// NormalRecur indicates an expected call of NormalRecur.
func (mr *MockClientMockRecorder) NormalRecur(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NormalRecur", reflect.TypeOf((*MockClient)(nil).NormalRecur), arg0, arg1, arg2)
}

// TooManyArgs mocks base method.
func (m *MockClient) TooManyArgs(arg0 context.Context, arg1 map[string]string, arg2 *bar.Bar_TooManyArgs_Args) (context.Context, *bar.BarResponse, map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TooManyArgs", arg0, arg1, arg2)
	ret0, _ := ret[0].(context.Context)
	ret1, _ := ret[1].(*bar.BarResponse)
	ret2, _ := ret[2].(map[string]string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// TooManyArgs indicates an expected call of TooManyArgs.
func (mr *MockClientMockRecorder) TooManyArgs(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TooManyArgs", reflect.TypeOf((*MockClient)(nil).TooManyArgs), arg0, arg1, arg2)
}
