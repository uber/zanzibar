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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"

	zanzibar "github.com/uber/zanzibar/runtime"
)

var testDir = getDirName()

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

func TestEmptyConfig(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(nil, nil)

	config.SetSeedOrDie("k", "v")
	assert.Equal(t, config.MustGetString("k"), "v")

	config.SetSeedOrDie("k2", true)
	assert.Equal(t, config.MustGetBoolean("k2"), true)

	config.SetSeedOrDie("k3", int64(4))
	assert.Equal(t, config.MustGetInt("k3"), int64(4))

	config.SetSeedOrDie("k4", float64(4.0))
	assert.Equal(t, config.MustGetFloat("k4"), float64(4.0))

	assert.Equal(t, config.ContainsKey("k5"), false)
	config.SetSeedOrDie("k5", "xyz")
	assert.Equal(t, config.ContainsKey("k5"), true)
}

func TestGetNamespace(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(nil, nil)

	config.SetSeedOrDie("a.b.c", "v")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")
}

func TestPanicNonExistantKeys(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(nil, nil)

	assert.Panics(t, func() {
		config.MustGetString("a.b.c")
	})

	assert.Panics(t, func() {
		config.MustGetBoolean("a.b.c")
	})

	assert.Panics(t, func() {
		config.MustGetInt("a.b.c")
	})

	assert.Panics(t, func() {
		config.MustGetFloat("a.b.c")
	})

	assert.Panics(t, func() {
		config.MustGetStruct("a.b.c", nil)
	})
}

func TestPanicGetWrongTypes(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(
		[]*zanzibar.ConfigOption{},
		map[string]interface{}{
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.2),
			"json":  `"{a":"b"}`,
		},
	)

	assert.Panics(t, func() {
		config.MustGetBoolean("int")
	})

	assert.Panics(t, func() {
		config.MustGetFloat("bool")
	})

	assert.Panics(t, func() {
		config.MustGetInt("a.b.c")
	})

	assert.Panics(t, func() {
		config.MustGetInt("float")
	})

	assert.Panics(t, func() {
		config.MustGetString("bool")
	})

	assert.Panics(t, func() {
		var x bool
		config.MustGetStruct("json", &x)
	})
}

func TestSupportsSeedConfig(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(
		[]*zanzibar.ConfigOption{},
		map[string]interface{}{
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.0),
			"exist": "xyz",
		},
	)

	assert.Equal(t, config.MustGetString("a.b.c"), "v")
	assert.Equal(t, config.MustGetBoolean("bool"), true)
	assert.Equal(t, config.MustGetInt("int"), int64(1))
	assert.Equal(t, config.MustGetInt("float"), int64(1))
	assert.Equal(t, config.MustGetFloat("float"), float64(1.0))
	assert.Equal(t, config.ContainsKey("exist"), true)
}

func TestCannotSetExistingKeys(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(
		[]*zanzibar.ConfigOption{},
		map[string]interface{}{
			"a.b.c": "v",
		},
	)

	assert.Panics(t, func() {
		config.SetSeedOrDie("a.b.c", "x")
	})
}

func TestCannotGetFromDestroyedConfig(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(
		[]*zanzibar.ConfigOption{},
		map[string]interface{}{
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.0),
			"exist": "xyz",
		},
	)

	assert.Equal(t, config.MustGetString("a.b.c"), "v")
	assert.Equal(t, config.MustGetBoolean("bool"), true)
	assert.Equal(t, config.MustGetInt("int"), int64(1))
	assert.Equal(t, config.MustGetFloat("float"), float64(1.0))

	config.Destroy()

	assert.Panics(t, func() {
		config.MustGetString("a.b.c")
	})

	assert.Panics(t, func() {
		assert.Equal(t, config.MustGetBoolean("bool"), true)
	})

	assert.Panics(t, func() {
		assert.Equal(t, config.MustGetInt("int"), int64(1))
	})

	assert.Panics(t, func() {
		assert.Equal(t, config.MustGetFloat("float"), float64(1.0))
	})

	assert.Panics(t, func() {
		var x bool
		config.MustGetStruct("bool", &x)
		assert.Equal(t, x, true)
	})

	assert.Panics(t, func() {
		assert.Equal(t, config.ContainsKey("exist"), true)
	})
}

func TestCannotSetOnFrozenConfig(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(nil, nil)

	config.SetSeedOrDie("a", "a")
	config.SetSeedOrDie("b", "b")

	config.Freeze()

	assert.Panics(t, func() {
		config.SetSeedOrDie("c", "c")
	})

	assert.Equal(t, config.InspectOrDie(), map[string]interface{}{
		"a": "a",
		"b": "b",
	})
}

func TestSetConfigValue(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(nil, nil)

	config.SetConfigValueOrDie("a", []byte("a"), "string")
	config.SetConfigValueOrDie("b", []byte("1"), "number")
	config.SetConfigValueOrDie("c", []byte("true"), "boolean")

	assert.Panics(t, func() {
		// wrong type
		config.SetConfigValueOrDie("d", []byte("unknown"), "unknown")
	})

	assert.Panics(t, func() {
		// parsing err
		config.SetConfigValueOrDie("d", []byte("d"), "number")
	})

	config.Freeze()

	assert.Panics(t, func() {
		// set after frozen
		config.SetConfigValueOrDie("d", []byte("d"), "string")
	})

	assert.Equal(t, config.InspectOrDie(), map[string]interface{}{
		"a": "a",
		"b": float64(1),
		"c": true,
	})
}

func mustMarshalJSON(v interface{}) []byte {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return bytes
}

func mustMarshalYAML(v interface{}) []byte {
	bytes, err := yaml.Marshal(v)
	if err != nil {
		panic(err)
	}

	return bytes
}

type fixtureWriter struct {
	rootDir  string
	fixtures map[string][]byte
}

func WriteFixture(rootDir string, fixtures map[string][]byte) *fixtureWriter {
	writer := &fixtureWriter{
		rootDir:  rootDir,
		fixtures: fixtures,
	}

	writer.writeToDisk()

	return writer
}

func (writer *fixtureWriter) writeToDisk() {
	for fileName, bytes := range writer.fixtures {
		filePath := filepath.Join(writer.rootDir, fileName)

		err := os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(filePath, bytes, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func (writer *fixtureWriter) Close() {
	for fileName := range writer.fixtures {
		parts := strings.Split(fileName, "/")
		filePath := filepath.Join(writer.rootDir, parts[0])

		err := os.RemoveAll(filePath)
		if err != nil {
			panic(err)
		}
	}
}

func DoCanReadFromFileTest(t *testing.T, filePath string) {
	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(filePath),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "b")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")
	assert.Equal(t, config.MustGetBoolean("bool"), true)
	assert.Equal(t, config.MustGetInt("int"), int64(1))
	assert.Equal(t, config.ContainsKey("exist"), true)

	type testStruct struct {
		Boolean bool
		Integer int64
		Float   float64
		String  string
	}
	var actual testStruct
	expected := testStruct{
		Boolean: true,
		Integer: 1,
		Float:   1.2,
		String:  "Science",
	}
	config.MustGetStruct("struct", &actual)
	assert.Equal(t, expected, actual)

	var actualArray []string
	expectedArray := []string{"a", "b"}
	config.MustGetStruct("array", &actualArray)
	assert.Equal(t, expectedArray, actualArray)

	actualMap := make(map[string]int64)
	expectedMap := map[string]int64{"key1": 10, "key2": 100, "key3": 10000, "key4": 9999999}
	config.MustGetStruct("mapStringInt", &actualMap)
	assert.Equal(t, expectedMap, actualMap)
}

func TestCanReadJSONFromFile(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.json": mustMarshalJSON(map[string]interface{}{
			"a":     "b",
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.2),
			"exist": "xyz",
			"struct": map[string]interface{}{
				"Boolean": true,
				"Integer": int64(1),
				"Float":   float64(1.2),
				"String":  "Science",
			},
			"array":        []string{"a", "b"},
			"mapStringInt": map[string]int64{"key1": 10, "key2": 100, "key3": 10000, "key4": 9999999},
		}),
	})
	DoCanReadFromFileTest(t, filepath.Join(testDir, "config", "test.json"))
	closer.Close()
}

func TestCanReadYAMLFromFile(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]interface{}{
			"a":     "b",
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.2),
			"exist": "xyz",
			"struct": map[string]interface{}{
				"Boolean": true,
				"Integer": int64(1),
				"Float":   float64(1.2),
				"String":  "Science",
			},
			"array":        []string{"a", "b"},
			"mapStringInt": map[string]int64{"key1": 10, "key2": 100, "key3": 10000, "key4": 9999999},
		}),
	})
	DoCanReadFromFileTest(t, filepath.Join(testDir, "config", "test.yaml"))
	closer.Close()
}

func TestCanReadFromFileContents(t *testing.T) {
	bytes := mustMarshalYAML(map[string]interface{}{
		"a":     "b",
		"a.b.c": "v",
		"bool":  true,
		"int":   int64(1),
		"float": float64(1.2),
		"exist": "xyz",
		"struct": map[string]interface{}{
			"Boolean": true,
			"Integer": int64(1),
			"Float":   float64(1.2),
			"String":  "Science",
		},
		"array": []string{"a", "b"},
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFileContents(bytes),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "b")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")
	assert.Equal(t, config.MustGetBoolean("bool"), true)
	assert.Equal(t, config.MustGetInt("int"), int64(1))
	assert.Equal(t, config.MustGetFloat("float"), float64(1.2))
	assert.Equal(t, config.ContainsKey("exist"), true)

	type testStruct struct {
		Boolean bool
		Integer int64
		Float   float64
		String  string
	}
	var actual testStruct
	expected := testStruct{
		Boolean: true,
		Integer: 1,
		Float:   1.2,
		String:  "Science",
	}
	config.MustGetStruct("struct", &actual)
	assert.Equal(t, expected, actual)

	var actualArray []string
	expectedArray := []string{"a", "b"}
	config.MustGetStruct("array", &actualArray)
	assert.Equal(t, expectedArray, actualArray)
}

func TestCannotSetOverValueFromFile(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]interface{}{
			"a":     "b",
			"a.b.c": "v",
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
	}, nil)

	assert.Panics(t, func() {
		config.SetSeedOrDie("a", "c")
	})

	closer.Close()
}

func TestCannotSetOverValueFromFileContents(t *testing.T) {
	bytes := mustMarshalYAML(map[string]interface{}{
		"a":     "b",
		"a.b.c": "v",
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFileContents(bytes),
	}, nil)

	assert.Panics(t, func() {
		config.SetSeedOrDie("a", "c")
	})
}

func TestReadFromSeedConfigIntoNil(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie(nil, map[string]interface{}{
		"a": "c",
	})

	assert.Panics(t, func() {
		config.MustGetStruct("a", nil)
	})
}

func TestDecodeIncompatibleStruct(t *testing.T) {
	bytes := mustMarshalYAML(map[string]interface{}{
		"struct": map[string]interface{}{
			"Boolean": true,
		},
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFileContents(bytes),
	}, nil)

	var obj []string
	assert.Panics(t, func() {
		config.MustGetStruct("struct", &obj)
	})
}

func TestSeedConfigOverwritesFiles(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
	}, map[string]interface{}{
		"a": "c",
	})

	assert.Equal(t, config.MustGetString("a"), "c")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")

	closer.Close()
}

func TestLaterFilesOverwriteEarlierFiles(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
		"config/local.yaml": mustMarshalYAML(map[string]string{
			"a": "c",
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "local.yaml"),
		),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "c")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")

	closer.Close()
}

func TestLaterContentsOverwriteEarlierFiles(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
	})

	bytes := mustMarshalYAML(map[string]string{
		"a": "c",
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
		zanzibar.ConfigFileContents(bytes),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "c")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")

	closer.Close()
}

func TestLaterFilesOverwriteEarlierContents(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/local.yaml": mustMarshalYAML(map[string]string{
			"a": "c",
		}),
	})

	bytes := mustMarshalYAML(map[string]string{
		"a":     "b",
		"a.b.c": "v",
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFileContents(bytes),
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "local.yaml"),
		),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "c")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")

	closer.Close()
}

func TestSupportsNonExistantFiles(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "local.json"),
		),
	}, map[string]interface{}{
		"a":     "d",
		"a.b.c": "v2",
	})

	assert.Equal(t, config.MustGetString("a"), "d")
	assert.Equal(t, config.MustGetString("a.b.c"), "v2")
}

func TestThrowsForInvalidFile(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
		"config/local.json": []byte("{{"),
	})

	assert.Panics(t, func() {
		_ = zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
			zanzibar.ConfigFilePath(
				filepath.Join(testDir, "config", "test.yaml"),
			),
			zanzibar.ConfigFilePath(
				filepath.Join(testDir, "config", "local.json"),
			),
		}, nil)
	})

	closer.Close()
}

func TestThrowsForInvalidContents(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
	})

	assert.Panics(t, func() {
		_ = zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
			zanzibar.ConfigFilePath(
				filepath.Join(testDir, "config", "test.yaml"),
			),
			zanzibar.ConfigFileContents([]byte("{{{")),
		}, nil)
	})

	closer.Close()
}

func TestThrowsForReadingBadFiles(t *testing.T) {
	assert.Panics(t, func() {
		_ = zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
			zanzibar.ConfigFilePath(filepath.Join(testDir)),
		}, nil)
	})
}

func TestGetStructFromDisk(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]interface{}{
			"a": map[string]interface{}{
				"Field": "a",
				"Foo":   4,
			},
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
	}, nil)

	var m struct {
		Field string
		Foo   int
	}
	config.MustGetStruct("a", &m)
	assert.Equal(t, m.Field, "a")
	assert.Equal(t, m.Foo, 4)

	closer.Close()
}

func TestInspectMalformedDataFromDisk(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": []byte(
			`{ "a": { "c": %%% }, "b": "c" }`,
		),
	})

	assert.Panics(t, func() {
		zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
			zanzibar.ConfigFilePath(
				filepath.Join(testDir, "config", "test.yaml"),
			),
		}, nil)
	})

	closer.Close()
}

func TestReadStructIntoWrongType(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]interface{}{
			"a":     true,
			"array": []string{"x", "y"},
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
	}, nil)

	var m struct {
		Field string
		Foo   int
	}
	assert.Panics(t, func() {
		config.MustGetStruct("a", &m)
	})

	array := []int{}
	assert.Panics(t, func() {
		config.MustGetStruct("a", &array)
	})

	assert.Panics(t, func() {
		config.MustGetStruct("array", &array)
	})

	assert.Panics(t, func() {
		config.MustGetStruct("unknown", &m)
	})

	closer.Close()
}

func TestOverwriteStructFromSeedConfig(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]interface{}{
			"a": map[string]interface{}{
				"Field": "a",
				"Foo":   4,
			},
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
	}, map[string]interface{}{
		"a": struct {
			Field string
			Foo   int
		}{
			Field: "b",
			Foo:   5,
		},
	})

	var m struct {
		Field string
		Foo   int
	}
	config.MustGetStruct("a", &m)
	assert.Equal(t, m.Field, "b")
	assert.Equal(t, m.Foo, 5)

	closer.Close()
}

func TestInspectFromFile(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]interface{}{
			"a":     "b",
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.2),
			"struct": map[string]interface{}{
				"Field": "a",
				"Foo":   4,
			},
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
	}, nil)

	assert.Equal(t, config.InspectOrDie(), map[string]interface{}{
		"a":     "b",
		"a.b.c": "v",
		"bool":  true,
		"float": float64(1.2),
		"int":   float64(1),
		"struct": map[string]interface{}{
			"Field": "a",
			"Foo":   float64(4),
		},
	})

	closer.Close()
}

func TestPanicReadingWrongTypeFromDisk(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/test.yaml": mustMarshalYAML(map[string]interface{}{
			"a":     "b",
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.2),
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "test.yaml"),
		),
	}, nil)

	assert.Panics(t, func() {
		config.MustGetBoolean("a")
	})

	assert.Panics(t, func() {
		config.MustGetString("int")
	})

	assert.Panics(t, func() {
		config.MustGetInt("a.b.c")
	})

	assert.Panics(t, func() {
		config.MustGetInt("float")
	})

	assert.Panics(t, func() {
		config.MustGetFloat("bool")
	})

	closer.Close()
}

func TestStaticConfigHasOwnState(t *testing.T) {
	dict := map[string]interface{}{
		"a.b.c": "v",
		"bool":  true,
		"int":   int64(1),
		"float": float64(1.2),
	}

	config1 := zanzibar.NewStaticConfigOrDie(
		[]*zanzibar.ConfigOption{},
		dict,
	)
	config2 := zanzibar.NewStaticConfigOrDie(
		[]*zanzibar.ConfigOption{},
		dict,
	)

	config1.SetSeedOrDie("a-key", "a-value")

	assert.Panics(t, func() {
		config2.MustGetString("a-key")
	})
}

func TestStaticConfigPanicBadConfig(t *testing.T) {
	assert.Panics(t, func() {
		_ = zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
			{},
		}, nil)
	})
}

func TestReadFromDict(t *testing.T) {
	testConfig := zanzibar.NewStaticConfigOrDie(nil, map[string]interface{}{
		"intFromDict":   500,
		"floatFromDict": 1,
	})
	intActual := int(testConfig.MustGetInt("intFromDict"))
	assert.Equal(t, 500, intActual)
	floatActual := testConfig.MustGetFloat("floatFromDict")
	assert.Equal(t, float64(1), floatActual)
}

func TestAsYaml(t *testing.T) {
	payload := map[string]interface{}{
		"a": 500,
		"b": "sd",
		"c": map[string]interface{}{
			"d": false,
			"e": -13,
		},
	}
	testCases := []struct {
		wantFrozen       bool
		wantDestroyed    bool
		wantErrorString  string
		payloadCfg       map[string]interface{}
		wantNonEmptyYaml bool
	}{
		{
			wantFrozen:       true,
			wantDestroyed:    false,
			wantNonEmptyYaml: true,
			payloadCfg:       payload,
		}, {
			wantFrozen:      true,
			wantDestroyed:   true,
			wantErrorString: "error representing as YAML, config is destroyed",
			payloadCfg:      payload,
		}, {
			wantFrozen:      false,
			wantDestroyed:   false,
			wantErrorString: "error representing as YAML, config is not frozen yet",
			payloadCfg:      payload,
		}, {
			wantFrozen:      true,
			wantDestroyed:   false,
			wantErrorString: "error representing as YAML, failed to serialize values",
			payloadCfg:      map[string]interface{}{"x": make(chan int)},
		},
	}
	for _, tc := range testCases {
		cfg := zanzibar.NewStaticConfigOrDie(nil, tc.payloadCfg)
		if tc.wantFrozen {
			cfg.Freeze()
		}
		if tc.wantDestroyed {
			cfg.Destroy()
		}
		asYaml, err := cfg.AsYaml()
		if tc.wantErrorString != "" {
			assert.Error(t, err, tc.wantErrorString)
		}
		if tc.wantNonEmptyYaml {
			assert.True(t, len(asYaml) != 0)
		}
	}
}
