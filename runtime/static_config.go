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
	"io/ioutil"
	"reflect"

	"encoding/json"

	"os"

	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
)

// StaticConfigValue represents a json serialized string.
type StaticConfigValue struct {
	bytes    []byte
	dataType jsonparser.ValueType
}

// StaticConfig allows accessing values of out of json config files
type StaticConfig struct {
	seedConfig   map[string]interface{}
	files        []string
	configValues map[string]StaticConfigValue
	frozen       bool
	destroyed    bool
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
	files []string, seedConfig map[string]interface{},
) *StaticConfig {
	if seedConfig == nil {
		seedConfig = map[string]interface{}{}
	}

	config := &StaticConfig{
		files:        files,
		seedConfig:   seedConfig,
		configValues: map[string]StaticConfigValue{},
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
			panic(errors.Wrapf(err, "Key (%s) is wrong type: ", key))
		}

		return v
	}

	panic(errors.Errorf("Key (%s) not available", key))
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
		rv := reflect.ValueOf(ptr)
		if rv.Kind() != reflect.Ptr || rv.IsNil() {
			panic(errors.Errorf("Cannot GetStruct (%s) into nil ptr", key))
		}

		rv.Set(reflect.ValueOf(v))
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

// SetOrDie a value in the config, useful for tests.
// Keys you set must not exist in the JSON files
// Set() will panic if the key exists or if frozen.
// Strongly recommended not to be used for production code.
func (conf *StaticConfig) SetOrDie(key string, value interface{}) {
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

	for i := 0; i < len(conf.files); i++ {
		fileObject := conf.parseFile(conf.files[i])
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

func (conf *StaticConfig) parseFile(fileName string) map[string]StaticConfigValue {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			// Ignore missing files
			return nil
		}

		// If the ReadFile() failed then just panic out.
		panic(err)
	}

	var object = map[string]StaticConfigValue{}

	err = jsonparser.ObjectEach(bytes, func(
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
