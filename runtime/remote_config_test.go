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
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
)

const tmpCfgPath = "/tmp/zanzibar_remote_cfg.json"
const refreshInterval = time.Second * 10
const defaultCfgJsonStr = `[
		{
			"key": "cfgboolean",
			"value": true,
		},
		{
			"key": "cfgint",
			"value": 20
		},
		{
			"key": "cfgfloat",
			"value": 1.5
		},
		{
			"key": "cfgstring",
			"value": "testStr"
		},
		{
			"key": "cfgstruct",
			"value": {
				"NestedStruct":{
					"ValBoolean": true,
					"ValInt": 48,
					"ValFloat": 3.1415926,
					"ValString": "testNestedStructStr"
				},
				"ValBoolean":true,
				"ValInt": 32,
				"ValFloat": 3.14,
				"ValString": "testStructStr"
			}
		}]`

type testSuite struct {
	configFilePath string
	configStr      string
	remoteConfig   zanzibar.RemoteConfig
}

type testStruct struct {
	NestedStruct *testNestedStruct
	ValBoolean   bool
	ValInt       int64
	ValFloat     float64
	ValString    string
}

type testNestedStruct struct {
	ValBoolean bool
	ValInt     int64
	ValFloat   float64
	ValString  string
}

func setupRemoteConfigTestSuite(filePath string, configStr string, interval time.Duration) *testSuite {
	configFileStr := defaultCfgJsonStr
	configFilePath := tmpCfgPath
	pollingInternal := refreshInterval
	if len(filePath) > 0 {
		configFilePath = filePath
	}
	if len(configStr) > 0 {
		configFileStr = configStr
	}
	if interval > 0 {
		pollingInternal = interval
	}
	err := ioutil.WriteFile(configFilePath, []byte(configFileStr), 0644)
	if err != nil {
		panic("unable to setup remote config file in " + configFilePath)
	}
	cfg := &zanzibar.RemoteConfigOptions{
		FilePath:        configFilePath,
		PollingInterval: pollingInternal,
	}
	remoteConfig, err := zanzibar.NewRemoteConfig(cfg, zap.NewNop(), tally.NoopScope)
	if err != nil {
		panic("unable to init new remote config")
	}
	return &testSuite{
		configFilePath: configFilePath,
		configStr:      configFileStr,
		remoteConfig:   remoteConfig,
	}
}

func (ts *testSuite) tearDown() {
	_ = os.Remove(ts.configFilePath)
	ts.remoteConfig.Close()
}

func TestNilFileInitializeError(t *testing.T) {
	rc, err := zanzibar.NewRemoteConfig(&zanzibar.RemoteConfigOptions{}, zap.NewNop(), nil)
	assert.Nil(t, rc)
	assert.NotNil(t, err)
	rc, err = zanzibar.NewRemoteConfig(&zanzibar.RemoteConfigOptions{}, nil, tally.NoopScope)
	assert.Nil(t, rc)
	assert.NotNil(t, err)
	rc, err = zanzibar.NewRemoteConfig(&zanzibar.RemoteConfigOptions{}, zap.NewNop(), tally.NoopScope)
	assert.Nil(t, rc)
	assert.NotNil(t, err)
}

func TestInvalidPathInitializeError(t *testing.T) {
	rc, err := zanzibar.NewRemoteConfig(
		&zanzibar.RemoteConfigOptions{FilePath: "invalid_path", PollingInterval: time.Second},
		zap.NewNop(),
		tally.NoopScope,
	)
	assert.Nil(t, rc)
	assert.NotNil(t, err)
}

func TestInvalidJSJONInitializeError(t *testing.T) {
	configFileStr := "{\\invalid%@;{ json"
	err := ioutil.WriteFile(tmpCfgPath, []byte(configFileStr), 0644)
	assert.Nil(t, err)
	defer func() {
		_ = os.Remove(tmpCfgPath)
	}()
	cfg := &zanzibar.RemoteConfigOptions{
		FilePath:        tmpCfgPath,
		PollingInterval: time.Second,
	}
	remoteConfig, err := zanzibar.NewRemoteConfig(cfg, zap.NewNop(), tally.NoopScope)
	assert.Nil(t, remoteConfig)
	assert.NotNil(t, err)
}

func TestInvalidConfigItem(t *testing.T) {
	configFileStr := `["not-object", {"missing-key": 1}, {"key":"test", "missing-value":1}]`
	err := ioutil.WriteFile(tmpCfgPath, []byte(configFileStr), 0644)
	assert.Nil(t, err)
	defer func() {
		_ = os.Remove(tmpCfgPath)
	}()
	cfg := &zanzibar.RemoteConfigOptions{
		FilePath:        tmpCfgPath,
		PollingInterval: time.Second,
	}
	remoteConfig, err := zanzibar.NewRemoteConfig(cfg, zap.NewNop(), tally.NoopScope)
	assert.NotNil(t, remoteConfig)
	assert.Nil(t, err)
	res := remoteConfig.GetString("missing-key", "")
	assert.Equal(t, "", res)
	res = remoteConfig.GetString("test", "")
	assert.Equal(t, "", res)
}

func TestTypeMisMatch(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", `[{"key":"cfgboolean", "value":true}]`, 0)
	defer ts.tearDown()
	vf := ts.remoteConfig.GetFloat("cfgboolean", float64(0))
	assert.Equal(t, float64(0), vf)
}

func TestUnmarshalErrStruct(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", `[{"key":"cfgstruct","value":{"NestedStruct":123}}]`, 0)
	defer ts.tearDown()
	ns := &testStruct{}
	ok := ts.remoteConfig.GetStruct("cfgstruct", ns)
	assert.False(t, ok)
}

func TestTypeMismatchNumber(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	vf := ts.remoteConfig.GetFloat("cfgint", float64(0))
	assert.Equal(t, float64(20), vf)
	vi := ts.remoteConfig.GetInt("cfgfloat", int64(0))
	assert.Equal(t, int64(0), vi)
}

func TestConcurrentRW(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	go func() {
		_ = ts.remoteConfig.Refresh()
	}()
	go func() {
		v := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
		assert.Equal(t, float64(1.5), v)
	}()
}

func TestConcurrentR(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	f := func() {
		v := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
		assert.Equal(t, float64(1.5), v)
	}
	go f()
	go f()
}

func TestConcurrentW(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	var wg sync.WaitGroup
	wg.Add(2)
	err := ioutil.WriteFile(tmpCfgPath, []byte(`[{"key":"test","value":1}]`), 0644)
	if err != nil {
		panic("unable to setup remote config file in " + tmpCfgPath)
	}
	f := func() {
		err = ts.remoteConfig.Refresh()
		assert.Nil(t, err)
		wg.Done()
	}
	go f()
	go f()
	wg.Wait()
	res := ts.remoteConfig.GetInt("test", int64(0))
	assert.True(t, res == int64(1))
}

func TestRefresh(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	vf := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
	assert.Equal(t, float64(1.5), vf)
	configFileStr := `[{"key":"cfgfloat", "value": 3.5}]`
	err := ioutil.WriteFile(tmpCfgPath, []byte(configFileStr), 0644)
	assert.Nil(t, err)
	err = ts.remoteConfig.Refresh()
	assert.Nil(t, err)
	vf = ts.remoteConfig.GetFloat("cfgfloat", float64(0))
	assert.Equal(t, float64(3.5), vf)
}

func TestTypeRemoteConfigAndDefaultFallback(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "[]", 0)
	defer ts.tearDown()
	vb := ts.remoteConfig.GetBoolean("cfgboolean", false)
	assert.Equal(t, false, vb)

	vi := ts.remoteConfig.GetInt("cfgint", int64(0))
	assert.Equal(t, int64(0), vi)

	vf := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
	assert.Equal(t, float64(0), vf)

	vs := ts.remoteConfig.GetString("cfgstring", "")
	assert.Equal(t, "", vs)

	teststruct := &testStruct{NestedStruct: &testNestedStruct{}}
	err := ts.remoteConfig.GetStruct("cfgstruct", teststruct)
	assert.NotNil(t, err)
}

func TestTypeHappyRemoteConfig(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()

	vb := ts.remoteConfig.GetBoolean("cfgboolean", false)
	assert.Equal(t, true, vb)

	vi := ts.remoteConfig.GetInt("cfgint", int64(0))
	assert.Equal(t, int64(20), vi)

	vf := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
	assert.Equal(t, float64(1.5), vf)

	vs := ts.remoteConfig.GetString("cfgstring", "")
	assert.Equal(t, "testStr", vs)

	teststruct := &testStruct{NestedStruct: &testNestedStruct{}}
	ok := ts.remoteConfig.GetStruct("cfgstruct", teststruct)
	expstruct := &testStruct{
		NestedStruct: &testNestedStruct{
			ValBoolean: true,
			ValInt:     int64(48),
			ValFloat:   float64(3.1415926),
			ValString:  "testNestedStructStr",
		},
		ValInt:     int64(32),
		ValBoolean: true,
		ValFloat:   float64(3.14),
		ValString:  "testStructStr",
	}
	assert.True(t, ok)
	assert.EqualValues(t, expstruct, teststruct)
}

func TestSubscribe(t *testing.T) {
	var (
		wg sync.WaitGroup
		f1 zanzibar.Callback
		f2 zanzibar.Callback
	)
	ts := setupRemoteConfigTestSuite("", "", 0)
	ch := make(chan int, 2)
	defer ts.tearDown()
	wg.Add(2)
	concurrentSub := func(identifier, key string, fn *zanzibar.Callback) {
		ts.remoteConfig.Subscribe(identifier, key, fn)
		wg.Done()
	}
	f1 = func(map[string]bool) { ch <- 1 }
	f2 = func(map[string]bool) { ch <- 2 }
	go concurrentSub("subscriber1", "cfgstring", &f1)
	go concurrentSub("subscriber2", "cfgint", &f2)
	wg.Wait()
	err := ioutil.WriteFile(tmpCfgPath, []byte(`[{"key":"cfgint","value":20}]`), 0644)
	assert.Nil(t, err)
	err = ts.remoteConfig.Refresh()
	assert.Nil(t, err)
	assert.Equal(t, 1, <-ch)
}

func TestUnsubscribe(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	ch := make(chan int, 1)
	defer ts.tearDown()
	var fn zanzibar.Callback
	fn = func(map[string]bool) {
		ch <- 1
	}
	updateNRefresh := func() {
		err := ioutil.WriteFile(tmpCfgPath, []byte("[]"), 0644)
		assert.Nil(t, err)
		err = ts.remoteConfig.Refresh()
		assert.Nil(t, err)
	}
	ts.remoteConfig.Subscribe("subscriber", "cfgstring", &fn)
	updateNRefresh()
	assert.Equal(t, 1, <-ch)
	ts.remoteConfig.Unsubscribe("subscriber")
	updateNRefresh()
	select {
	case <-ch:
		assert.Fail(t, "unsubscriber event should not emit")
	default:
		return
	}
}

func TestSubscribeNoSpecificKey(t *testing.T) {
	var fn zanzibar.Callback
	ts := setupRemoteConfigTestSuite("", "", 0)
	ch := make(chan int, 1)
	defer ts.tearDown()
	fn = func(diff map[string]bool) {
		ch <- 1
		assert.Equal(t, 4, len(diff))
	}
	ts.remoteConfig.Subscribe("subscriber", "", &fn)
	err := ioutil.WriteFile(tmpCfgPath, []byte(`[{"key":"cfgint","value":20}]`), 0644)
	assert.Nil(t, err)
	err = ts.remoteConfig.Refresh()
	assert.Nil(t, err)
	assert.Equal(t, 1, <-ch)
}

func TestPolling(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", time.Millisecond)
	defer ts.tearDown()
	err := ioutil.WriteFile(tmpCfgPath, []byte(`[{"key":"test", "value":1}]`), 0644)
	assert.Nil(t, err)
	time.Sleep(2 * time.Millisecond)
	res := ts.remoteConfig.GetInt("test", int64(0))
	assert.Equal(t, int64(1), res)
}

func TestPollingNilFile(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", time.Millisecond)
	defer ts.tearDown()
	_ = os.Remove(tmpCfgPath)
	time.Sleep(2 * time.Millisecond)
	res := ts.remoteConfig.GetInt("test", int64(0))
	assert.Equal(t, int64(0), res)
}

func TestRefreshNilFile(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	_ = os.Remove(tmpCfgPath)
	err := ts.remoteConfig.Refresh()
	assert.NotNil(t, err)
}
