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
	"testing"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
)

func TestGet(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}
	zh["foo"] = []string{"headOne"}
	zh["bar"] = []string{"otherHeader"}

	v, ok := zh.Get("foo")
	assert.Equal(t, true, ok)
	assert.Equal(t, "headOne", v)
}

func TestGetOrEmpty(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}
	zh["foo"] = []string{"headOne"}
	zh["bar"] = []string{""}

	v := zh.GetOrEmptyStr("foo")
	assert.Equal(t, "headOne", v)

	v = zh.GetOrEmptyStr("missing")
	assert.Equal(t, "", v)

	v = zh.GetOrEmptyStr("bar")
	assert.Equal(t, "", v)
}

func TestGetMissingKey(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}

	v, ok := zh.Get("foo")
	assert.Equal(t, false, ok)
	assert.Equal(t, "", v)
}

func TestGetMultivalueKey(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}
	zh["foo"] = []string{"headOne", "headTwo"}
	zh["bar"] = []string{"otherHeader"}

	v, ok := zh.Get("foo")
	assert.Equal(t, true, ok)
	assert.Equal(t, "headOne", v)
}

func TestAdd(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}
	zh["bar"] = []string{"otherHeader"}

	zh.Add("foo", "headOne")
	assert.Equal(t, "headOne", zh["foo"][0])
}

func TestAddMultivalueKey(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}
	zh["foo"] = []string{"headOne"}
	zh["bar"] = []string{"otherHeader"}

	zh.Add("foo", "headTwo")
	assert.Equal(t, "headOne", zh["foo"][0])
	assert.Equal(t, "headTwo", zh["foo"][1])
}

func TestSetOverwriteOldKey(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}
	zh["foo"] = []string{"headOne"}
	zh["bar"] = []string{"otherHeader"}

	zh.Set("foo", "newHeader")
	assert.Equal(t, "newHeader", zh["foo"][0])
	assert.Equal(t, 1, len(zh["foo"]))
}

func TestSetNewKey(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}
	zh["bar"] = []string{"otherHeader"}

	zh.Set("foo", "headOne")
	assert.Equal(t, "headOne", zh["foo"][0])
}

func TestSetOverwriteMultiKey(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}
	zh["foo"] = []string{"headOne", "headTwo"}
	zh["bar"] = []string{"otherHeader"}

	zh.Set("foo", "newHeader")
	assert.Equal(t, "newHeader", zh["foo"][0])
	assert.Equal(t, 1, len(zh["foo"]))
}

func TestMissingKey(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}
	emptyAr := make([]string, 0)
	assert.Equal(t, emptyAr, zh.Keys())
}

func TestKeys(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}

	zh["foo"] = []string{"headOne"}
	zh["bar"] = []string{"otherHeader"}

	assert.Equal(t, 2, len(zh.Keys()))
}

func TestEnsure(t *testing.T) {
	zh := zanzibar.ServerHTTPHeader{}

	zh["foo"] = []string{"headOne"}
	zh["bar"] = []string{"otherHeader"}

	assert.Equal(t, nil, zh.Ensure([]string{"foo"}))
	assert.Error(t, zh.Ensure([]string{"foo", "baz"}))
}

func TestSTHGet(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}
	zh["foo"] = "headOne"

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
	zh["foo"] = "headOne"

	zh.Set("foo", "newHeader")
	assert.Equal(t, "newHeader", zh["foo"])
}

func TestSTHSetNewKey(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}
	zh["bar"] = "otherHeader"

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

	zh["foo"] = "headOne"
	zh["bar"] = "otherHeader"

	assert.Equal(t, 2, len(zh.Keys()))
}

func TestSTHEnsure(t *testing.T) {
	zh := zanzibar.ServerTChannelHeader{}

	zh["foo"] = "headOne"
	zh["bar"] = "otherHeader"

	assert.Equal(t, nil, zh.Ensure([]string{"foo"}))
	assert.Error(t, zh.Ensure([]string{"foo", "baz"}))
}
