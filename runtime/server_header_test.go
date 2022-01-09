// Copyright (c) 2022 Uber Technologies, Inc.
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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
)

func TestGet(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	zh.Set("foo", "headOne")
	zh.Set("bar", "otherHeader")

	v, ok := zh.Get("foo")
	assert.Equal(t, true, ok)
	assert.Equal(t, "headOne", v)
}

func TestGetOrEmpty(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	zh.Set("foo", "headOne")
	zh.Set("bar", "")

	v := zh.GetOrEmptyStr("foo")
	assert.Equal(t, "headOne", v)

	v = zh.GetOrEmptyStr("missing")
	assert.Equal(t, "", v)

	v = zh.GetOrEmptyStr("bar")
	assert.Equal(t, "", v)
}

func TestGetMissingKey(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})

	v, ok := zh.Get("foo")
	assert.Equal(t, false, ok)
	assert.Equal(t, "", v)
}

func TestGetMultivalueKey(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	zh.Set("foo", "headOne")
	zh.Add("foo", "headTwo")
	zh.Set("bar", "otherHeader")

	v, ok := zh.Get("foo")
	assert.Equal(t, true, ok)
	assert.Equal(t, "headOne", v)
}

func TestValues(t *testing.T) {
	key := "Canonicalized-Key"
	testCases := []struct {
		title          string
		header         func() zanzibar.ServerHTTPHeader
		expectedValues []string
		expectedBool   bool
	}{
		{
			title: "Multiple values for a valid key",
			header: func() zanzibar.ServerHTTPHeader {
				zh := zanzibar.NewServerHTTPHeader(http.Header{})
				zh.Set(key, "headerOne")
				zh.Add(key, "headerTwo")
				return zh
			},
			expectedValues: []string{"headerOne", "headerTwo"},
			expectedBool:   true,
		},
		{
			title: "Single value for a valid key",
			header: func() zanzibar.ServerHTTPHeader {
				zh := zanzibar.NewServerHTTPHeader(http.Header{})
				zh.Set(key, "headerOne")
				return zh
			},
			expectedValues: []string{"headerOne"},
			expectedBool:   true,
		},
		{
			title: "Single value containing comma-separated inner values for a valid key",
			header: func() zanzibar.ServerHTTPHeader {
				zh := zanzibar.ServerHTTPHeader{}
				zh.Set(key, "headerOne,headerTwo")
				return zh
			},
			expectedValues: []string{"headerOne,headerTwo"},
			expectedBool:   true,
		},
		{
			title: "Zero values for a valid key",
			header: func() zanzibar.ServerHTTPHeader {
				zh := zanzibar.NewServerHTTPHeader(http.Header{
					key: []string{},
				})
				return zh
			},
			expectedValues: []string{},
			expectedBool:   true,
		},
		{
			title: "Missing header key",
			header: func() zanzibar.ServerHTTPHeader {
				zh := zanzibar.NewServerHTTPHeader(http.Header{})
				return zh
			},
			expectedValues: []string{},
			expectedBool:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			actualValues, actualBool := tc.header().Values(key)
			assert.Equal(t, tc.expectedValues, actualValues, tc.title)
			assert.Equal(t, tc.expectedBool, actualBool, tc.title)
		})
	}
}

func TestAdd(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	zh.Set("bar", "otherHeader")

	zh.Add("foo", "headOne")
	assert.Equal(t, "headOne", zh.GetOrEmptyStr("foo"))
}

func TestAddMultivalueKey(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	zh.Set("foo", "headOne")
	zh.Set("bar", "otherHeader")

	zh.Add("foo", "headTwo")
	assert.Equal(t,
		[]string{"headOne", "headTwo"},
		zh.GetAll("foo"),
	)
}

func TestSetOverwriteOldKey(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	zh.Set("foo", "headOne")
	zh.Set("bar", "otherHeader")

	zh.Set("foo", "newHeader")
	assert.Equal(t, "newHeader", zh.GetOrEmptyStr("foo"))
	assert.Equal(t, 1, len(zh.GetAll("foo")))
}

func TestSetNewKey(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	zh.Set("bar", "otherHeader")

	zh.Set("foo", "headOne")
	assert.Equal(t, "headOne", zh.GetOrEmptyStr("foo"))
}

func TestUnsetKey(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	zh.Set("foo", "bar")
	assert.Equal(t, "bar", zh.GetOrEmptyStr("foo"))

	zh.Unset("foo")
	v, ok := zh.Get("foo")
	assert.False(t, ok)
	assert.Equal(t, "", v)
}

func TestSetOverwriteMultiKey(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	zh.Set("foo", "headOne")
	zh.Add("foo", "headTwo")
	zh.Set("bar", "otherHeader")

	zh.Set("foo", "newHeader")
	assert.Equal(t, "newHeader", zh.GetOrEmptyStr("foo"))
	assert.Equal(t, 1, len(zh.GetAll("foo")))
}

func TestMissingKey(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})
	emptyAr := make([]string, 0)
	assert.Equal(t, emptyAr, zh.Keys())
}

func TestKeys(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})

	zh.Set("foo", "headOne")
	zh.Set("bar", "otherHeader")

	assert.Equal(t, 2, len(zh.Keys()))
}

func TestEnsure(t *testing.T) {
	zh := zanzibar.NewServerHTTPHeader(http.Header{})

	zh.Set("foo", "headOne")
	zh.Set("bar", "otherHeader")

	assert.Equal(t, nil, zh.Ensure([]string{"foo"}, zap.NewNop()))
	assert.Error(t, zh.Ensure([]string{"foo", "baz"}, zap.NewNop()))
}

func TestSTHGet(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}
	zh.Set("foo", "headOne")

	v, ok := zh.Get("foo")
	assert.Equal(t, true, ok)
	assert.Equal(t, "headOne", v)
}

func TestSTHGetMissingKey(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}

	v, ok := zh.Get("foo")
	assert.Equal(t, false, ok)
	assert.Equal(t, "", v)
}

func TestSTHValues(t *testing.T) {
	key := "Canonicalized-Key"
	testCases := []struct {
		title          string
		header         func() zanzibar.ServerTChannelHeader
		expectedValues []string
		expectedBool   bool
	}{
		{
			title: "Multiple values set for a valid key",
			header: func() zanzibar.ServerTChannelHeader {
				zh := zanzibar.ServerTChannelHeader{}
				zh.Set(key, "headerOne")
				// For ServerTChannelHeader, Add is an alias to Set so
				// this will overwrite the existing key with the new value.
				zh.Add(key, "headerTwo")
				return zh
			},
			expectedValues: []string{"headerTwo"},
			expectedBool:   true,
		},
		{
			title: "Single value for a valid key",
			header: func() zanzibar.ServerTChannelHeader {
				zh := zanzibar.ServerTChannelHeader{}
				zh.Set(key, "headerOne")
				return zh
			},
			expectedValues: []string{"headerOne"},
			expectedBool:   true,
		},
		{
			title: "Single value containing comma-separated inner values for a valid key",
			header: func() zanzibar.ServerTChannelHeader {
				zh := zanzibar.ServerTChannelHeader{}
				zh.Set(key, "headerOne,headerTwo")
				return zh
			},
			expectedValues: []string{"headerOne,headerTwo"},
			expectedBool:   true,
		},
		{
			title: "Missing header key",
			header: func() zanzibar.ServerTChannelHeader {
				zh := zanzibar.ServerTChannelHeader{}
				return zh
			},
			expectedValues: []string{},
			expectedBool:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.title, func(t *testing.T) {
			actualValues, actualBool := tc.header().Values(key)
			assert.Equal(t, tc.expectedValues, actualValues, tc.title)
			assert.Equal(t, tc.expectedBool, actualBool, tc.title)
		})
	}
}

func TestSTHAdd(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}

	zh.Add("foo", "headOne")
	v, ok := zh.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, "headOne", v)
}

func TestSTHSetOverwriteOldKey(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}
	zh.Set("foo", "headOne")

	zh.Add("foo", "newHeader")
	v, ok := zh.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, "newHeader", v)
}

func TestSTHSetNewKey(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}
	zh.Set("bar", "otherHeader")

	zh.Add("foo", "headOne")
	v, ok := zh.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, "headOne", v)
}

func TestSTHUnset(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}
	zh.Add("foo", "bar")
	v, ok := zh.Get("foo")
	assert.True(t, ok)
	assert.Equal(t, "bar", v)

	zh.Unset("foo")
	v, ok = zh.Get("foo")
	assert.False(t, ok)
	assert.Equal(t, "", v)
}

func TestSTHMissingKey(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}
	emptyAr := make([]string, 0)
	assert.Equal(t, emptyAr, zh.Keys())
}

func TestSTHkeys(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}

	zh.Set("foo", "headOne")
	zh.Set("bar", "otherHeader")

	assert.Equal(t, 2, len(zh.Keys()))
}

func TestSTHEnsure(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}

	zh.Set("foo", "headOne")
	zh.Set("bar", "otherHeader")

	assert.Equal(t, nil, zh.Ensure([]string{"foo"}, zap.NewNop()))
	assert.Error(t, zh.Ensure([]string{"foo", "baz"}, zap.NewNop()))
}
