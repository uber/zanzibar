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

package zanzibar

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"

	"github.com/ghodss/yaml"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

type configType int

const (
	filePathConfigType     configType = 1
	fileContentsConfigType configType = 2
)

// StaticConfig allows accessing values out of YAML(JSON) config files
type StaticConfig struct {
	seedConfig    map[string]interface{}
	configOptions []*ConfigOption
	configValues  map[string]interface{}
	frozen        bool
	destroyed     bool
}

// ConfigOption points to a collection of bytes representing a file. This is
// either the full contents of the file as bytes, or a string poiting to a
// file
type ConfigOption struct {
	configType configType
	bytes      []byte
}

// ConfigFilePath creates a ConfigFile represented as a path
func ConfigFilePath(path string) *ConfigOption {
	return &ConfigOption{
		configType: filePathConfigType,
		bytes:      []byte(path),
	}
}

// ConfigFileContents creates a ConfigFile representing the contents of the file
func ConfigFileContents(fileBytes []byte) *ConfigOption {
	return &ConfigOption{
		configType: fileContentsConfigType,
		bytes:      fileBytes,
	}
}

// NewStaticConfigOrDie allocates a static config instance
// StaticConfig takes two args, files and seedConfig.
//
// files is required and must be a list of file paths.
// The later files overwrite keys from earlier files.
//
// The seedConfig is optional and will be used to overwrite
// configuration files if present.
//
// The defaultConfig is optional initial config that will be overwritten by
// the config in the supplied files or the seed config
//
// The files must be a list of YAML files. Each file must be a flat object of
// key, value pairs. It's recommended that you use keys like:
//
// {
//     "name": "my-name",
//     "clients.thingy": {
//          "some client": "config"
//     },
//     "server.my-port": 9999
// }
//
// To organize your configuration file.
func NewStaticConfigOrDie(
	configOptions []*ConfigOption,
	seedConfig map[string]interface{},
) *StaticConfig {
	config := &StaticConfig{
		configOptions: configOptions,
		seedConfig:    map[string]interface{}{},
		configValues:  map[string]interface{}{},
	}

	for key, value := range seedConfig {
		config.seedConfig[key] = value
	}

	config.initializeConfigValues()

	return config
}

func (conf *StaticConfig) checkConfDestroyed(key string) {
	if conf.destroyed {
		panic(errors.Errorf("cannot get(%s) because destroyed", key))
	}
}

// MustGetBoolean returns the value as a boolean or panics.
func (conf *StaticConfig) MustGetBoolean(key string) bool {
	conf.checkConfDestroyed(key)

	if value, contains := conf.seedConfig[key]; contains {
		return value.(bool)
	}

	if value, contains := conf.configValues[key]; contains {
		return value.(bool)
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

func mustConvertableToFloat(value interface{}, key string) float64 {
	if v, ok := value.(int); ok {
		return float64(v)
	}
	return value.(float64)
}

// MustGetFloat returns the value as a float or panics.
func (conf *StaticConfig) MustGetFloat(key string) float64 {
	conf.checkConfDestroyed(key)

	if value, contains := conf.seedConfig[key]; contains {
		return mustConvertableToFloat(value, key)
	}

	if value, contains := conf.configValues[key]; contains {
		return mustConvertableToFloat(value, key)
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

func mustConvertableToInt(value interface{}, key string) int64 {
	if v, ok := value.(float64); ok {
		// value can be represented as an integer
		if v != float64(int64(v)) {
			panic(errors.Errorf("Key (%s) is a float", key))
		}
		return int64(v)
	}
	if v, ok := value.(int); ok {
		return int64(v)
	}
	return value.(int64)
}

// MustGetInt returns the value as a int or panics.
func (conf *StaticConfig) MustGetInt(key string) int64 {
	conf.checkConfDestroyed(key)

	if value, contains := conf.seedConfig[key]; contains {
		return mustConvertableToInt(value, key)
	}

	if value, contains := conf.configValues[key]; contains {
		return mustConvertableToInt(value, key)
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

// ContainsKey returns true if key is found otherwise false.
func (conf *StaticConfig) ContainsKey(key string) bool {
	if conf.destroyed {
		panic(errors.Errorf("Cannot ContainsKey(%s) because destroyed", key))
	}
	if _, contains := conf.seedConfig[key]; contains {
		return true
	}

	if _, contains := conf.configValues[key]; contains {
		return true
	}
	return false
}

// MustGetString returns the value as a string or panics.
func (conf *StaticConfig) MustGetString(key string) string {
	conf.checkConfDestroyed(key)

	if value, contains := conf.seedConfig[key]; contains {
		return value.(string)
	}

	if value, contains := conf.configValues[key]; contains {
		return value.(string)
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

// MustGetStruct reads the value into an interface{} or panics.
// Recommended that this is used with pointers to structs
func (conf *StaticConfig) MustGetStruct(key string, ptr interface{}) {
	conf.checkConfDestroyed(key)

	rptr := reflect.ValueOf(ptr)
	if rptr.Kind() != reflect.Ptr || rptr.IsNil() {
		panic(errors.Errorf("Cannot GetStruct (%s) into nil ptr", key))
	}
	if v, contains := conf.seedConfig[key]; contains {
		rptr.Elem().Set(reflect.ValueOf(v))
		return
	}

	if v, contains := conf.configValues[key]; contains {
		err := conf.mustGetStructHelper(key, v, ptr)
		if err != nil {
			panic(err)
		}
		return
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

func (conf *StaticConfig) mustGetStructHelper(key string, v interface{}, ptr interface{}) error {
	if value, ok := v.(map[string]interface{}); ok {
		bytes, err := json.Marshal(value)
		if err != nil {
			panic(errors.Errorf("Decoding key (%s) failed", key))
		}
		err = json.Unmarshal(bytes, ptr)
		if err != nil {
			panic(errors.Errorf("Decoding key (%s) failed", key))
		}
		return nil
	} else if value, ok := v.([]interface{}); ok {
		err := mapstructure.Decode(value, ptr)
		if err != nil {
			panic(errors.Errorf("Decoding key (%s) failed", key))
		}
		return nil
	}

	return errors.Errorf("cant cast value into one of known types (%s)", key)
}

// SetConfigValueOrDie sets the static config value.
// dataType can be a boolean, number or string.
// SetConfigValueOrDie will panic if the config is frozen.
func (conf *StaticConfig) SetConfigValueOrDie(
	key string, bytes []byte, dataType string) {
	if conf.frozen {
		panic(errors.Errorf("Cannot set(%s) because frozen", key))
	}

	var value interface{}
	var err error
	switch dataType {
	case "boolean":
		value, err = strconv.ParseBool(string(bytes))
	case "number":
		value, err = strconv.ParseFloat(string(bytes), 64)
	case "string":
		value, err = string(bytes), nil
	default:
		panic("unknown config data type")
	}

	if err != nil {
		panic(errors.Errorf("Parsing config value (%s) falsed", string(bytes)))
	}

	conf.configValues[key] = value
}

// SetSeedOrDie a value in the config, useful for tests.
// Keys you set must not exist in the YAML files.
// Set() will panic if the key exists or if frozen.
// Strongly recommended not to be used for production code.
func (conf *StaticConfig) SetSeedOrDie(key string, value interface{}) {
	if conf.frozen {
		panic(errors.Errorf("Cannot set(%s) because frozen", key))
	}

	if _, contains := conf.configValues[key]; contains {
		panic(errors.Errorf("Key (%s) already exists", key))
	}

	if _, contains := conf.seedConfig[key]; contains {
		panic(errors.Errorf("Key (%s) already exists", key))
	}

	conf.seedConfig[key] = value
}

// Freeze the configuration store.
// Once you freeze the config any further calls to config.set() will panic.
// This allows you to make the static config immutable
func (conf *StaticConfig) Freeze() {
	conf.frozen = true
}

// Destroy will make Get() calls fail with a panic.
// This allows you to terminate the configuration phase and gives you
// confidence that your application is now officially bootstrapped.
func (conf *StaticConfig) Destroy() {
	conf.destroyed = true
	conf.frozen = true
	conf.configValues = map[string]interface{}{}
	conf.seedConfig = map[string]interface{}{}
}

// InspectOrDie returns the entire config object.
// This should not be mutated and should only be used for inspection or debugging
func (conf *StaticConfig) InspectOrDie() map[string]interface{} {
	result := map[string]interface{}{}

	for k, v := range conf.configValues {
		result[k] = v
	}

	for k, v := range conf.seedConfig {
		result[k] = v
	}
	return result
}

// AsYaml returns a YAML serialized version of the StaticConfig
// If the StaticConfig is destroyed or frozen, this method returns an error
func (conf *StaticConfig) AsYaml() ([]byte, error) {
	if conf.destroyed {
		return nil, errors.New("error representing as YAML, config is destroyed")
	}
	if !conf.frozen {
		return nil, errors.New("error representing as YAML, config is not frozen yet. Use Freeze() to mark the config as frozen")
	}
	yamlBytes, err := yaml.Marshal(conf.InspectOrDie())
	if err != nil {
		return nil, errors.Wrap(err, "error representing as YAML, failed to serialize values")
	}
	return yamlBytes, nil
}

func (conf *StaticConfig) initializeConfigValues() {
	values := conf.collectConfigMaps()
	conf.assignConfigValues(values)
}

func (conf *StaticConfig) collectConfigMaps() []map[string]interface{} {
	maps := make([]map[string]interface{}, len(conf.configOptions))

	for i := 0; i < len(conf.configOptions); i++ {
		fileObject := conf.parseFile(conf.configOptions[i])
		if fileObject != nil {
			maps = append(maps, fileObject)
		}
	}

	return maps
}

func (conf *StaticConfig) assignConfigValues(values []map[string]interface{}) {
	for i := 0; i < len(values); i++ {
		configObject := values[i]

		for key, value := range configObject {
			conf.configValues[key] = value
		}
	}
}

func (conf *StaticConfig) parseFile(
	configFile *ConfigOption,
) map[string]interface{} {
	var bytes []byte

	switch configFile.configType {
	case filePathConfigType:
		var err error
		bytes, err = ioutil.ReadFile(string(configFile.bytes))
		if err != nil {
			if os.IsNotExist(err) {
				// Ignore missing files
				return nil
			}

			// If the ReadFile() failed then just panic out.
			panic(err)
		}
	case fileContentsConfigType:
		bytes = configFile.bytes
	default:
		panic(errors.Errorf(
			"Unknown config file type %d",
			configFile.configType,
		))
	}

	var object map[string]interface{}
	err := yaml.Unmarshal(bytes, &object)
	if err != nil {
		panic(err)
	}
	return object
}
