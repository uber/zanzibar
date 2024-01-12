// Copyright (c) 2024 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zanzibar_test

import (
	"context"
	"fmt"
	"net/textproto"
	"testing"

	"github.com/stretchr/testify/assert"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	zanzibar "github.com/uber/zanzibar/runtime"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	"go.uber.org/thriftrw/wire"
)

func TestCreatingTChannel(t *testing.T) {
	_, err := benchGateway.CreateGateway(
		map[string]interface{}{
			"tchannel.serviceName": "",
			"serviceName":          "",
		},
		nil,
		exampleGateway.CreateGateway,
	)
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "no service name provided")
}

func TestDecorateWithRecover(t *testing.T) {
	testCases := map[string]struct {
		handleFn     func(context.Context, *wire.Value) (zanzibar.RWTStruct, error)
		expectedResp zanzibar.RWTStruct
		expectedErr  error
	}{
		"Success: decorate returns response when no panic occurs": {
			handleFn: func(context.Context, *wire.Value) (zanzibar.RWTStruct, error) {
				return mockRWTStruct{}, nil
			},
			expectedResp: mockRWTStruct{},
			expectedErr:  nil,
		},
		"Error: decorate returns error when error without panic occurs": {
			handleFn: func(context.Context, *wire.Value) (zanzibar.RWTStruct, error) {
				return nil, fmt.Errorf("handle function failed")
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("handle function failed"),
		},
		"Error: decorate returns error when panic occurs": {
			handleFn: func(context.Context, *wire.Value) (zanzibar.RWTStruct, error) {
				panic("handle function fails")
			},
			expectedResp: nil,
			expectedErr:  fmt.Errorf("panic: handle function fails"),
		},
	}

	for tc, tt := range testCases {
		t.Run(tc, func(t *testing.T) {
			resp, err := zanzibar.DecorateWithRecover(context.Background(), nil, nil, tt.handleFn)
			assert.Equal(t, tt.expectedResp, resp)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

type mockRWTStruct struct{}

func (m mockRWTStruct) ToWire() (wire.Value, error) {
	return wire.Value{}, nil
}

func (m mockRWTStruct) FromWire(wire.Value) error {
	return nil
}

func TestGetRequestUUID(t *testing.T) {
	uuidHeaderKey := "test-uuid-key"
	testUuidValue := uuid.New()

	testCases := map[string]struct {
		reqHeaders     map[string]string
		expectedEquals string
	}{
		"Success: return new UUID when no request header map is given": {},
		"Success: return new UUID when no request header map is given": {
			reqHeaders: map[string]string{},
		},
		"Success: return uuid when provided through the headerMap with any casing": {
			reqHeaders: map[string]string{
				uuidHeaderKey: testUuidValue,
			},
			expectedEquals: testUuidValue,
		},
		"Success: return a UUID when multiple are provided through the headerMap": {
			reqHeaders: map[string]string{
				uuidHeaderKey:   testUuidValue,
				"TEST-uuid-KEY": testUuidValue,
				textproto.CanonicalMIMEHeaderKey(uuidHeaderKey): testUuidValue,
			},
			expectedEquals: testUuidValue,
		},
	}

	for tc, tt := range testCases {
		t.Run(tc, func(t *testing.T) {
			resp, err := zanzibar.DecorateWithRecover(context.Background(), nil, nil, tt.handleFn)
			uuid := getRequestUUID(uuidHeaderKey, tc.reqHeaders)

			assert.NotEqual(t, "", uuid, "request UUID should generate a new UUID if one does not exist")
			if tt.expectedEquals != "" {
				assert.Equal(t, tt.expectedEquals, uuid, "request UUID should be returned if one does exist")
			}
		})
	}
}
