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
	"bytes"
	"encoding/json"
	"fmt"
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
	GetBoolean(string, bool) bool
	GetFloat(string, float64) float64
	GetInt(string, int64) int64
	GetString(string, string) string
	GetStruct(string, interface{}) bool
	Subscribe(string, string, *func())
	Unsubscribe(string)
	Refresh() error
	Close()
}

// RemoteConfigOptions configs remote config file path and refresh frequency
type RemoteConfigOptions struct {
	FilePath        string
	PollingInterval time.Duration
}

// Subscriber is used to track subscribers' callbacks and keys they subscribe
type Subscriber struct {
	Callback *func()
	Key      string
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
	subscribers map[string]*Subscriber
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
// remote config file follows the same requirements with runtime/static_config/StaticConfig
// remote config consider fallback legit case, fallback happens silent without throwing error
// Why RemoteConfig:
// 		What diverge staticConfig with remoteConfig is that staticConfig allows value setting
//		while remoteConfig is read only.
//		remoteConfig also provides pub/sub to allow caller act on specific key changes
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
		subscribers: make(map[string]*Subscriber),
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
	return rc.props.Load().(RemoteConfigMap)
}

func (rc *remoteConfig) getValidatedValue(key string, vt jsonparser.ValueType) (*RemoteConfigValue, bool) {
	ret, ok := rc.loadConfig()[key]
	if !ok {
		rc.logger.Warn(fmt.Sprintf("Key <%s> missing", key))
		return nil, false
	}
	if ret.dataType != vt {
		rc.logger.Warn(fmt.Sprintf("Key <%s> type mismatch (expected %s actual %s)", key, vt, ret.dataType))
		return nil, false
	}
	return ret, true
}

// GetBoolean returns the value as a boolean
func (rc *remoteConfig) GetBoolean(key string, fallback bool) bool {
	if ret, ok := rc.getValidatedValue(key, jsonparser.Boolean); ok {
		v, _ := jsonparser.ParseBoolean(ret.bytes)
		return v
	}
	return fallback
}

// GetFloat returns the value as a float64
func (rc *remoteConfig) GetFloat(key string, fallback float64) float64 {
	if ret, ok := rc.getValidatedValue(key, jsonparser.Number); ok {
		v, _ := jsonparser.ParseFloat(ret.bytes)
		return v
	}
	return fallback
}

// GetInt returns the value as int64
func (rc *remoteConfig) GetInt(key string, fallback int64) int64 {
	ret, ok := rc.getValidatedValue(key, jsonparser.Number)
	if !ok {
		return fallback
	}
	if v, err := jsonparser.ParseInt(ret.bytes); err == nil {
		return v
	}
	rc.logger.Warn(fmt.Sprintf("Key <%s> is not int64", key))
	return fallback
}

// GetString returns the value as string
func (rc *remoteConfig) GetString(key string, fallback string) string {
	if ret, ok := rc.getValidatedValue(key, jsonparser.String); ok {
		v, _ := jsonparser.ParseString(ret.bytes)
		return v
	}
	return fallback
}

// GetStruct loads struct ptr interface{}
func (rc *remoteConfig) GetStruct(key string, ptr interface{}) bool {
	ret, ok := rc.getValidatedValue(key, jsonparser.Object)
	if !ok {
		return false
	}
	if err := json.Unmarshal(ret.bytes, ptr); err != nil {
		return false
	}
	return true
}

func changedKeys(prev, curr RemoteConfigMap) map[string]bool {
	ret := make(map[string]bool)
	diff := func(a, b RemoteConfigMap) {
		for key, val := range a {
			if bval, ok := b[key]; ok && bytes.Compare(bval.bytes, val.bytes) == 0 && bval.dataType == val.dataType {
				continue
			}
			ret[key] = true
		}
	}
	diff(prev, curr)
	diff(curr, prev)
	return ret
}

// Refresh refreshes remote config and trigger callbacks if necessary.
// Refresh will check file stat to decide if reload in-mem config
// necessary. Reload config from file success will record keys
// been added/removed/updated, and trigger callbacks that registered
// via Subscribe.
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
	changedMap, err := rc.reloadConfigAndReturnChanged()
	if stat != nil && err == nil {
		rc.currStat = stat
	}
	rc.execCallbacks(changedMap)
	return err
}

// checkAndReturnStat returns file stat and boolean of `whether we should refresh from config file`
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

// reloadConfigAndReturnChanged atomically stores remote config properties
// and return keys been added/removed/updated
func (rc *remoteConfig) reloadConfigAndReturnChanged() (map[string]bool, error) {
	bytes, err := ioutil.ReadFile(rc.config.FilePath)
	changedMap := make(map[string]bool)
	if err != nil {
		return changedMap, err
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
	prevProps := rc.loadConfig()
	if err == nil {
		changedMap = changedKeys(prevProps, currProps)
		rc.props.Store(currProps)
	}
	return changedMap, err
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
func (rc *remoteConfig) Subscribe(identifier, key string, fn *func()) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.subscribers[identifier] = &Subscriber{
		Callback: fn,
		Key:      key,
	}
}

// Unsubscribe will stop an identifer's callback function from being executed upon
// config refresh
func (rc *remoteConfig) Unsubscribe(identifier string) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	delete(rc.subscribers, identifier)
}

func (rc *remoteConfig) execCallbacks(changedMap map[string]bool) {
	for _, s := range rc.subscribers {
		if _, ok := changedMap[s.Key]; ok {
			go (*s.Callback)()
		}
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
