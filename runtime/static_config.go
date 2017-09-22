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
	"reflect"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

type configType int

const (
	filePathConfigType     configType = 1
	fileContentsConfigType configType = 2
)

// StaticConfigValue represents a json serialized string.
type StaticConfigValue struct {
	bytes    []byte
	dataType jsonparser.ValueType
}

// StaticConfig allows accessing values out of json config files
type StaticConfig struct {
	seedConfig    map[string]interface{}
	configOptions []*ConfigOption
	configValues  map[string]StaticConfigValue
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
// configuration json files if present.
//
// The defaultConfig is optional initial config that will be overwritten by
// the config in the supplied files or the seed config
//
// The files must be a list of JSON files. Each file must be a flat object of
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
		configValues:  map[string]StaticConfigValue{},
	}

	for key, value := range seedConfig {
		config.seedConfig[key] = value
	}

	config.initializeConfigValues()

	return config
}

// MustGetBoolean returns the value as a boolean or panics.
func (conf *StaticConfig) MustGetBoolean(key string) bool {
	if conf.destroyed {
		panic(errors.Errorf("Cannot get(%s) because destroyed", key))
	}

	if value, contains := conf.seedConfig[key]; contains {
		return value.(bool)
	}

	if value, contains := conf.configValues[key]; contains {
		if value.dataType != jsonparser.Boolean {
			panic(errors.Errorf(
				"Key (%s) is not a boolean: %s", key, string(value.bytes),
			))
		}

		v, err := jsonparser.ParseBoolean(value.bytes)
		if err != nil {
			/* coverage ignore next line */
			panic(errors.Wrapf(err, "Key (%s) is wrong type: ", key))
		}

		return v
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

// MustGetFloat returns the value as a float or panics.
func (conf *StaticConfig) MustGetFloat(key string) float64 {
	if conf.destroyed {
		panic(errors.Errorf("Cannot get(%s) because destroyed", key))
	}

	if value, contains := conf.seedConfig[key]; contains {
		return value.(float64)
	}

	if value, contains := conf.configValues[key]; contains {
		if value.dataType != jsonparser.Number {
			panic(errors.Errorf(
				"Key (%s) is not a number: %s", key, string(value.bytes),
			))
		}

		v, err := jsonparser.ParseFloat(value.bytes)
		if err != nil {
			/* coverage ignore next line */
			panic(errors.Wrapf(err, "Key (%s) is wrong type: ", key))
		}

		return v
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

// MustGetInt returns the value as a int or panics.
func (conf *StaticConfig) MustGetInt(key string) int64 {
	if conf.destroyed {
		panic(errors.Errorf("Cannot get(%s) because destroyed", key))
	}

	if value, contains := conf.seedConfig[key]; contains {
		return value.(int64)
	}

	if value, contains := conf.configValues[key]; contains {
		if value.dataType != jsonparser.Number {
			panic(errors.Errorf(
				"Key (%s) is not a number: %s", key, string(value.bytes),
			))
		}

		v, err := jsonparser.ParseInt(value.bytes)
		if err != nil {
			/* coverage ignore next line */
			panic(errors.Wrapf(err, "Key (%s) is wrong type: ", key))
		}

		return v
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

// ContainsKey returns the value as a string or panics.
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
	if conf.destroyed {
		panic(errors.Errorf("Cannot get(%s) because destroyed", key))
	}

	if value, contains := conf.seedConfig[key]; contains {
		return value.(string)
	}

	if value, contains := conf.configValues[key]; contains {
		if value.dataType != jsonparser.String {
			panic(errors.Errorf(
				"Key (%s) is not a String: %s", key, string(value.bytes),
			))
		}

		v, err := jsonparser.ParseString(value.bytes)
		if err != nil {
			/* coverage ignore next line */
			panic(errors.Wrapf(err, "Key (%s) is wrong type: ", key))
		}

		return v
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

// MustGetStruct reads the value into an interface{} or panics.
// Recommended that this is used with pointers to structs
// MustGetStruct() will call json.Unmarshal(bytes, ptr) under the hood.
func (conf *StaticConfig) MustGetStruct(key string, ptr interface{}) {
	if conf.destroyed {
		panic(errors.Errorf("Cannot get(%s) because destroyed", key))
	}

	if v, contains := conf.seedConfig[key]; contains {
		rptr := reflect.ValueOf(ptr)
		if rptr.Kind() != reflect.Ptr || rptr.IsNil() {
			panic(errors.Errorf("Cannot GetStruct (%s) into nil ptr", key))
		}

		rptr.Elem().Set(reflect.ValueOf(v))
		return
	}

	if v, contains := conf.configValues[key]; contains {
		err := json.Unmarshal(v.bytes, ptr)
		if err != nil {
			panic(errors.Wrapf(err, "Key (%s) is wrong type: ", key))
		}

		return
	}

	panic(errors.Errorf("Key (%s) not available", key))
}

// SetConfigValueOrDie sets the static config value.
// dataType can be a boolean, number or string.
// SetConfigValueOrDie will panic if the config is frozen.
func (conf *StaticConfig) SetConfigValueOrDie(key string, bytes []byte, dataType string) {
	if conf.frozen {
		panic(errors.Errorf("Cannot set(%s) because frozen", key))
	}

	var dt jsonparser.ValueType
	switch dataType {
	case "boolean":
		dt = jsonparser.Boolean
	case "number":
		dt = jsonparser.Number
	case "string":
		dt = jsonparser.String
	default:
		panic("unknown config data type")
	}

	conf.configValues[key] = StaticConfigValue{
		dataType: dt,
		bytes:    bytes,
	}
}

// SetSeedOrDie a value in the config, useful for tests.
// Keys you set must not exist in the JSON files.
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
	conf.configValues = map[string]StaticConfigValue{}
	conf.seedConfig = map[string]interface{}{}
}

// InspectOrDie returns the entire config object.
// This should not be mutated and should only be used for inspection or debugging
func (conf *StaticConfig) InspectOrDie() map[string]interface{} {
	result := map[string]interface{}{}

	for k, v := range conf.configValues {
		var jsonValue interface{}
		var err error

		switch v.dataType {
		case jsonparser.Boolean:
			jsonValue, err = jsonparser.ParseBoolean(v.bytes)
		case jsonparser.String:
			jsonValue, err = jsonparser.ParseString(v.bytes)
		case jsonparser.Number:
			jsonValue, err = jsonparser.ParseFloat(v.bytes)
		default:
			err = json.Unmarshal(v.bytes, &jsonValue)
		}

		if err != nil {
			panic(errors.Wrapf(err, "Key (%s) is not json: ", k))
		}

		result[k] = jsonValue
	}

	for k, v := range conf.seedConfig {
		result[k] = v
	}

	return result
}

func (conf *StaticConfig) initializeConfigValues() {
	values := conf.collectConfigMaps()
	conf.assignConfigValues(values)
}

func (conf *StaticConfig) collectConfigMaps() []map[string]StaticConfigValue {
	var maps = []map[string]StaticConfigValue{}

	for i := 0; i < len(conf.configOptions); i++ {
		fileObject := conf.parseFile(conf.configOptions[i])
		if fileObject != nil {
			maps = append(maps, fileObject)
		}
	}

	return maps
}

func (conf *StaticConfig) assignConfigValues(values []map[string]StaticConfigValue) {
	for i := 0; i < len(values); i++ {
		configObject := values[i]

		for key, value := range configObject {
			conf.configValues[key] = value
		}
	}
}

func (conf *StaticConfig) parseFile(
	configFile *ConfigOption,
) map[string]StaticConfigValue {
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

	var object = map[string]StaticConfigValue{}

	err := jsonparser.ObjectEach(bytes, func(
		key []byte,
		value []byte,
		dataType jsonparser.ValueType,
		offset int,
	) error {
		object[string(key)] = StaticConfigValue{
			bytes:    value,
			dataType: dataType,
		}
		return nil
	})

	if err != nil {
		// If the JSON is not valid then just panic out.
		panic(err)
	}

	return object
}
