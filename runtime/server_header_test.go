// Copyright (c) 2017 Uber Technologies, Inc.
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

func TestSTHAdd(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}

	zh.Add("foo", "headOne")
	assert.Equal(t, "headOne", zh["foo"])
}

func TestSTHSetOverwriteOldKey(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}
	zh.Set("foo", "headOne")

	zh.Set("foo", "newHeader")
	assert.Equal(t, "newHeader", zh["foo"])
}

func TestSTHSetNewKey(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}
	zh.Set("bar", "otherHeader")

	zh.Set("foo", "headOne")
	assert.Equal(t, "headOne", zh["foo"])
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
