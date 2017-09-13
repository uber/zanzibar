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
	"io/ioutil"
	"os"
	"sync"
	"time"
)

const tmpCfgPath = "/tmp/zanzibar_remote_cfg.json"
const refreshInterval = time.Second * 10
const defaultCfgJsonStr = `{
		"cfgboolean": true,
		"cfgint": 20,
		"cfgfloat": 1.5,
		"cfgstring": "testStr",
		"cfgstruct": {
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
	}`

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
	remoteConfig, err := zanzibar.NewRemoteConfig(cfg, nil, nil)
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
	os.Remove(ts.configFilePath)
	ts.remoteConfig.Close()
}

func TestNilFileInitializeError(t *testing.T) {
	rc, err := zanzibar.NewRemoteConfig(&zanzibar.RemoteConfigOptions{}, nil, nil)
	assert.Nil(t, rc)
	assert.NotNil(t, err)
}

func TestInvalidPathInitializeError(t *testing.T) {
	rc, err := zanzibar.NewRemoteConfig(
		&zanzibar.RemoteConfigOptions{FilePath: "invalid_path", PollingInterval: time.Second},
		nil,
		nil,
	)
	assert.Nil(t, rc)
	assert.NotNil(t, err)
}

func TestInvalidJSJONInitializeError(t *testing.T) {
	configFileStr := "{\\invalid%@;{ json"
	err := ioutil.WriteFile(tmpCfgPath, []byte(configFileStr), 0644)
	defer os.Remove(tmpCfgPath)
	if err != nil {
		panic("unable to setup remote config file in " + tmpCfgPath)
	}
	cfg := &zanzibar.RemoteConfigOptions{
		FilePath:        tmpCfgPath,
		PollingInterval: time.Second,
	}
	remoteConfig, err := zanzibar.NewRemoteConfig(cfg, nil, nil)
	assert.Nil(t, remoteConfig)
	assert.NotNil(t, err)
}

func TestTypeMisMatch(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", `{"cfgboolean": true}`, 0)
	defer ts.tearDown()
	vf, err := ts.remoteConfig.GetFloat("cfgboolean", float64(0))
	assert.NotNil(t, err)
	assert.Equal(t, float64(0), vf)
}

func TestConcurrentRW(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	go ts.remoteConfig.Refresh()
	go func() {
		v, err := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
		assert.Equal(t, float64(1.5), v)
		assert.Nil(t, err)
	}()
}

func TestConcurrentR(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	f := func() {
		v, err := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
		assert.Equal(t, float64(1.5), v)
		assert.Nil(t, err)
	}
	go f()
	go f()
}

func TestConcurrentW(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	var wg sync.WaitGroup
	wg.Add(2)
	err := ioutil.WriteFile(tmpCfgPath, []byte(`{"test": 1}`), 0644)
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
	res, err := ts.remoteConfig.GetInt("test", int64(0))
	assert.Nil(t, err)
	assert.True(t, res == int64(1))
}

func TestRefresh(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	vf, err := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
	assert.Equal(t, float64(1.5), vf)
	assert.Nil(t, err)
	configFileStr := `{"cfgfloat": 3.5}`
	err = ioutil.WriteFile(tmpCfgPath, []byte(configFileStr), 0644)
	if err != nil {
		panic("unable to setup remote config file in " + tmpCfgPath)
	}
	err = ts.remoteConfig.Refresh()
	assert.Nil(t, err)
	vf, err = ts.remoteConfig.GetFloat("cfgfloat", float64(0))
	assert.Equal(t, float64(3.5), vf)
	assert.Nil(t, err)
}

func TestTypeRemoteConfigAndDefaultFallback(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "{}", 0)
	defer ts.tearDown()
	vb, err := ts.remoteConfig.GetBoolean("cfgboolean")
	assert.Equal(t, false, vb)
	assert.NotNil(t, err)

	vi, err := ts.remoteConfig.GetInt("cfgint", int64(0))
	assert.Equal(t, int64(0), vi)
	assert.NotNil(t, err)

	vf, err := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
	assert.Equal(t, float64(0), vf)
	assert.NotNil(t, err)

	vs, err := ts.remoteConfig.GetString("cfgstring", "")
	assert.Equal(t, "", vs)
	assert.NotNil(t, err)
}

func TestTypeHappyRemoteConfig(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()

	vb, err := ts.remoteConfig.GetBoolean("cfgboolean")
	assert.Equal(t, true, vb)
	assert.Nil(t, err)

	vi, err := ts.remoteConfig.GetInt("cfgint", int64(0))
	assert.Equal(t, int64(20), vi)
	assert.Nil(t, err)

	vf, err := ts.remoteConfig.GetFloat("cfgfloat", float64(0))
	assert.Equal(t, float64(1.5), vf)
	assert.Nil(t, err)

	vs, err := ts.remoteConfig.GetString("cfgstring", "")
	assert.Equal(t, "testStr", vs)
	assert.Nil(t, err)

	teststruct := &testStruct{NestedStruct: &testNestedStruct{}}
	err = ts.remoteConfig.GetStruct("cfgstruct", teststruct)
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
	assert.Nil(t, err)
	assert.EqualValues(t, expstruct, teststruct)
}

func TestSubscribe(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	ch := make(chan int, 2)
	defer ts.tearDown()
	var wg sync.WaitGroup
	wg.Add(2)
	concurrentSub := func(identifier string, fn *func()) {
		ts.remoteConfig.Subscribe(identifier, fn)
		wg.Done()
	}
	f1 := func() { ch <- 1 }
	f2 := func() { ch <- 1 }
	go concurrentSub("subscriber1", &f1)
	go concurrentSub("subscriber2", &f2)
	wg.Wait()
	err := ioutil.WriteFile(tmpCfgPath, []byte("{}"), 0644)
	if err != nil {
		panic("unable to setup remote config file in " + tmpCfgPath)
	}
	ts.remoteConfig.Refresh()
	for i := 0; i < 2; i++ {
		assert.Equal(t, 1, <-ch)
	}
}

func TestUnsubscribe(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	ch := make(chan int)
	defer ts.tearDown()
	fn := func() {
		ch <- 1
	}
	updateNRefresh := func() {
		err := ioutil.WriteFile(tmpCfgPath, []byte("{}"), 0644)
		if err != nil {
			panic("unable to setup remote config file in " + tmpCfgPath)
		}
		ts.remoteConfig.Refresh()
	}
	ts.remoteConfig.Subscribe("subscriber", &fn)
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

func TestPolling(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", time.Millisecond)
	defer ts.tearDown()
	err := ioutil.WriteFile(tmpCfgPath, []byte(`{"test": 1}`), 0644)
	if err != nil {
		panic("unable to setup remote config file in " + tmpCfgPath)
	}
	time.Sleep(2 * time.Millisecond)
	res, err := ts.remoteConfig.GetInt("test", int64(0))
	assert.Equal(t, int64(1), res)
}

func TestRefreshNilFile(t *testing.T) {
	ts := setupRemoteConfigTestSuite("", "", 0)
	defer ts.tearDown()
	os.Remove(tmpCfgPath)
	err := ts.remoteConfig.Refresh()
	assert.NotNil(t, err)
}
