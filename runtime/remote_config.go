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

package zanzibar

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

// RemoteConfig allow accessing values from a periodically refreshed json file
type RemoteConfig interface {
	GetBoolean(string) (bool, error)
	GetFloat(string, float64) (float64, error)
	GetInt(string, int64) (int64, error)
	GetString(string, string) (string, error)
	GetStruct(string, interface{}) error
	Subscribe(string, *func())
	Unsubscribe(string)
	Refresh() error
	Close()
}

// RemoteConfigOptions configs remote config file path and refresh frequency
type RemoteConfigOptions struct {
	FilePath        string
	PollingInterval time.Duration
}

// RemoteConfigValue represents a json serialized string
type RemoteConfigValue struct {
	bytes    []byte
	dataType jsonparser.ValueType
}

// RemoteConfigMap allows lock-free concurrent rw
// TODO consider https://golang.org/pkg/sync/#Map golang@1.9^
type RemoteConfigMap map[string]*RemoteConfigValue

type remoteConfig struct {
	config      *RemoteConfigOptions
	subscribers map[string]*func()
	props       atomic.Value
	close       chan struct{}
	currStat    os.FileInfo
	logger      *zap.Logger
	scope       tally.Scope
	poller      *time.Ticker
	mutex       sync.RWMutex
	wg          sync.WaitGroup
}

// NewRemoteConfig allocates a config that can be dynamically changed during runtime.
// RemoteConfigOptions takes two args, config file path (that remote config will read from)
// and polling interval (decides how frequent we check ).
// The remote config file follows the same requirements with runtime/static_config/StaticConfig
// Why RemoteConfig:
// 		What diverge staticConfig with remoteConfig is that staticConfig allows value setting
//		while remoteConfig is read only.
func NewRemoteConfig(cfg *RemoteConfigOptions, logger *zap.Logger, scope tally.Scope) (RemoteConfig, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	if len(cfg.FilePath) < 1 || cfg.PollingInterval == 0 {
		return nil, errors.Errorf("invalid remote config options")
	}
	if _, err := os.Stat(cfg.FilePath); os.IsNotExist(err) {
		return nil, errors.Errorf("invalid remote config file path")
	}
	if scope == nil {
		logger.Warn("no valid metrics scope")
		scope = tally.NoopScope
	}
	rc := &remoteConfig{
		config:      cfg,
		close:       make(chan struct{}, 1),
		subscribers: make(map[string]*func()),
		logger:      logger,
		scope:       scope,
		poller:      time.NewTicker(cfg.PollingInterval),
	}
	rc.props.Store(make(RemoteConfigMap))
	if err := rc.Refresh(); err != nil {
		return nil, err
	}
	rc.wg.Add(1)
	go rc.poll()
	return rc, nil
}

func (rc *remoteConfig) loadConfig() RemoteConfigMap {
	props := rc.props.Load()
	if props == nil {
		return nil
	}
	return props.(RemoteConfigMap)
}

func (rc *remoteConfig) getValidatedValue(key string, vt jsonparser.ValueType) (*RemoteConfigValue, error) {
	ret, ok := rc.loadConfig()[key]
	if !ok {
		return nil, errors.Errorf("Key <%s> does not exist", key)
	}
	if ret.dataType != vt {
		return nil, errors.Errorf("Key <%s> type mismatch (expected %s actual %s)", key, vt, ret.dataType)
	}
	return ret, nil
}

// GetBoolean returns the value as a boolean
func (rc *remoteConfig) GetBoolean(key string) (bool, error) {
	var (
		ret *RemoteConfigValue
		err error
	)
	if ret, err = rc.getValidatedValue(key, jsonparser.Boolean); err != nil {
		return false, err
	}
	if v, err := jsonparser.ParseBoolean(ret.bytes); err == nil {
		return v, nil
	}
	return false, errors.Errorf("Key <%s> is not boolean type", key)
}

// GetFloat returns the value as a float64
func (rc *remoteConfig) GetFloat(key string, fallback float64) (float64, error) {
	var (
		ret *RemoteConfigValue
		err error
	)
	if ret, err = rc.getValidatedValue(key, jsonparser.Number); err != nil {
		return fallback, err
	}
	if v, err := jsonparser.ParseFloat(ret.bytes); err == nil {
		return v, nil
	}
	return fallback, errors.Errorf("Key <%s> is not float64 type", key)
}

// GetInt returns the value as int64
func (rc *remoteConfig) GetInt(key string, fallback int64) (int64, error) {
	var (
		ret *RemoteConfigValue
		err error
	)
	if ret, err = rc.getValidatedValue(key, jsonparser.Number); err != nil {
		return fallback, err
	}
	if v, err := jsonparser.ParseInt(ret.bytes); err == nil {
		return v, nil
	}
	return fallback, errors.Errorf("Key <%s> is not int64 type", key)
}

// GetString returns the value as string
func (rc *remoteConfig) GetString(key string, fallback string) (string, error) {
	var (
		ret *RemoteConfigValue
		err error
	)
	if ret, err = rc.getValidatedValue(key, jsonparser.String); err != nil {
		return fallback, err
	}
	if v, err := jsonparser.ParseString(ret.bytes); err == nil {
		return v, nil
	}
	return fallback, errors.Errorf("Key <%s> is not string type", key)
}

// GetStruct loads struct ptr interface{}
func (rc *remoteConfig) GetStruct(key string, ptr interface{}) error {
	var (
		ret *RemoteConfigValue
		err error
	)
	if ret, err = rc.getValidatedValue(key, jsonparser.Object); err != nil {
		return err
	}
	return json.Unmarshal(ret.bytes, ptr)
}

// Refresh refreshes remote config if necessary.
func (rc *remoteConfig) Refresh() error {
	var (
		err  error
		stat os.FileInfo
	)
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	stat, refresh := rc.checkAndReturnStat()
	if !refresh {
		return nil
	}
	err = rc.reloadConfig()
	if stat != nil && err == nil {
		rc.currStat = stat
	}
	rc.execCallbacks()
	return err
}

// _checkAndReturnStat returns file stat and boolean of `whether we should refresh from config file`
// don't refresh if config file hasn't change since last refresh
func (rc *remoteConfig) checkAndReturnStat() (os.FileInfo, bool) {
	stat, err := os.Stat(rc.config.FilePath)
	if err != nil {
		rc.logger.Error("Error stat remote config file " + rc.config.FilePath)
		return nil, true
	}
	if rc.currStat != nil && rc.currStat.ModTime().Equal(stat.ModTime()) && rc.currStat.Size() == stat.Size() {
		return stat, false
	}
	return stat, true
}

func (rc *remoteConfig) reloadConfig() error {
	bytes, err := ioutil.ReadFile(rc.config.FilePath)
	if err != nil {
		return err
	}
	currProps := make(RemoteConfigMap)
	err = jsonparser.ObjectEach(bytes, func(
		key []byte,
		value []byte,
		dataType jsonparser.ValueType,
		offset int,
	) error {
		currProps[string(key)] = &RemoteConfigValue{
			bytes:    value,
			dataType: dataType,
		}
		return nil
	})
	if err == nil {
		rc.props.Store(currProps)
	}
	return err
}

// Close wait for all running refresh to complete, then shuts down the client
func (rc *remoteConfig) Close() {
	rc.poller.Stop()
	rc.close <- struct{}{}
	rc.wg.Wait()
}

// Subscribe adds a callback function to be executed when config refresh is complete
// Subscribe takes a string identifier and a function pointer to be called after
// a completed config refresh
func (rc *remoteConfig) Subscribe(identifier string, fn *func()) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.subscribers[identifier] = fn
}

// Unsubscribe will stop an identifer's callback function from being executed upon
// config refresh
func (rc *remoteConfig) Unsubscribe(identifier string) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	delete(rc.subscribers, identifier)
}

func (rc *remoteConfig) execCallbacks() {
	for _, fn := range rc.subscribers {
		go (*fn)()
	}
}

func (rc *remoteConfig) poll() {
	for {
		select {
		case <-rc.poller.C:
			result := "ok"
			if err := rc.Refresh(); err != nil {
				result = "err"
			}
			rc.scope.Tagged(map[string]string{"result": result}).Counter("remote_config.polling")
		case <-rc.close:
			rc.wg.Done()
			return
		}
	}
}
