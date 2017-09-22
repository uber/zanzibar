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
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"encoding/json"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
)

var testDir string = getDirName()

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
			"float": float64(1.0),
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
		config.MustGetString("bool")
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
		config.SetConfigValueOrDie("d", []byte("unknown"), "unknown")
	})

	config.Freeze()

	assert.Panics(t, func() {
		config.SetConfigValueOrDie("d", []byte("d"), "string")
	})

	assert.Equal(t, config.InspectOrDie(), map[string]interface{}{
		"a": "a",
		"b": float64(1),
		"c": true,
	})
}

func mustMarshal(v interface{}) []byte {
	bytes, err := json.Marshal(v)
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

func TestCanReadFromFile(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/production.json": mustMarshal(map[string]interface{}{
			"a":     "b",
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.0),
			"exist": "xyz",
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
		),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "b")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")
	assert.Equal(t, config.MustGetBoolean("bool"), true)
	assert.Equal(t, config.MustGetInt("int"), int64(1))
	assert.Equal(t, config.MustGetFloat("float"), float64(1.0))
	assert.Equal(t, config.ContainsKey("xyz"), true)

	closer.Close()
}

func TestCanReadFromFileContents(t *testing.T) {
	bytes := mustMarshal(map[string]interface{}{
		"a":     "b",
		"a.b.c": "v",
		"bool":  true,
		"int":   int64(1),
		"float": float64(1.0),
		"exist": "xyz",
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFileContents(bytes),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "b")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")
	assert.Equal(t, config.MustGetBoolean("bool"), true)
	assert.Equal(t, config.MustGetInt("int"), int64(1))
	assert.Equal(t, config.MustGetFloat("float"), float64(1.0))
	assert.Equal(t, config.ContainsKey("xyz"), true)
}

func TestCannotSetOverValueFromFile(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/production.json": mustMarshal(map[string]interface{}{
			"a":     "b",
			"a.b.c": "v",
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
		),
	}, nil)

	assert.Panics(t, func() {
		config.SetSeedOrDie("a", "c")
	})

	closer.Close()
}

func TestCannotSetOverValueFromFileContents(t *testing.T) {
	bytes := mustMarshal(map[string]interface{}{
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

func TestSeedConfigOverwritesFiles(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/production.json": mustMarshal(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
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
		"config/production.json": mustMarshal(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
		"config/local.json": mustMarshal(map[string]string{
			"a": "c",
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
		),
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "local.json"),
		),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "c")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")

	closer.Close()
}

func TestLaterContentsOverwriteEarlierFiles(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/production.json": mustMarshal(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
	})

	bytes := mustMarshal(map[string]string{
		"a": "c",
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
		),
		zanzibar.ConfigFileContents(bytes),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "c")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")

	closer.Close()
}

func TestLaterFilesOverwriteEarlierContents(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/local.json": mustMarshal(map[string]string{
			"a": "c",
		}),
	})

	bytes := mustMarshal(map[string]string{
		"a":     "b",
		"a.b.c": "v",
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFileContents(bytes),
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "local.json"),
		),
	}, nil)

	assert.Equal(t, config.MustGetString("a"), "c")
	assert.Equal(t, config.MustGetString("a.b.c"), "v")

	closer.Close()
}

func TestSupportsNonExistantFiles(t *testing.T) {
	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
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

func TestThrowsForInvalidJSONFile(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/production.json": mustMarshal(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
		"config/local.json": []byte("{...}"),
	})

	assert.Panics(t, func() {
		_ = zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
			zanzibar.ConfigFilePath(
				filepath.Join(testDir, "config", "production.json"),
			),
			zanzibar.ConfigFilePath(
				filepath.Join(testDir, "config", "local.json"),
			),
		}, nil)
	})

	closer.Close()
}

func TestThrowsForInvalidJSONContents(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/production.json": mustMarshal(map[string]string{
			"a":     "b",
			"a.b.c": "v",
		}),
	})

	assert.Panics(t, func() {
		_ = zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
			zanzibar.ConfigFilePath(
				filepath.Join(testDir, "config", "production.json"),
			),
			zanzibar.ConfigFileContents([]byte("{...}")),
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
		"config/production.json": mustMarshal(map[string]interface{}{
			"a": map[string]interface{}{
				"Field": "a",
				"Foo":   4,
			},
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
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
		"config/production.json": []byte(
			"{ \"a\": { \"c\": ... }, \"b\": \"c\" }",
		),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
		),
	}, nil)

	assert.Panics(t, func() {
		config.InspectOrDie()
	})

	closer.Close()
}

func TestReadStructIntoWrongType(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/production.json": mustMarshal(map[string]interface{}{
			"a": true,
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
		),
	}, nil)

	var m struct {
		Field string
		Foo   int
	}
	assert.Panics(t, func() {
		config.MustGetStruct("a", &m)
	})

	closer.Close()
}

func TestOverwriteStructFromSeedConfig(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/production.json": mustMarshal(map[string]interface{}{
			"a": map[string]interface{}{
				"Field": "a",
				"Foo":   4,
			},
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
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
		"config/production.json": mustMarshal(map[string]interface{}{
			"a":     "b",
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.0),
			"struct": map[string]interface{}{
				"Field": "a",
				"Foo":   4,
			},
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
		),
	}, nil)

	assert.Equal(t, config.InspectOrDie(), map[string]interface{}{
		"a":     "b",
		"a.b.c": "v",
		"bool":  true,
		"float": float64(1.0),
		"int":   float64(1),
		"struct": map[string]interface{}{
			"Field": "a",
			"Foo":   float64(4.0),
		},
	})

	closer.Close()
}

func TestPanicReadingWrongTypeFromDisk(t *testing.T) {
	closer := WriteFixture(testDir, map[string][]byte{
		"config/production.json": mustMarshal(map[string]interface{}{
			"a":     "b",
			"a.b.c": "v",
			"bool":  true,
			"int":   int64(1),
			"float": float64(1.0),
		}),
	})

	config := zanzibar.NewStaticConfigOrDie([]*zanzibar.ConfigOption{
		zanzibar.ConfigFilePath(
			filepath.Join(testDir, "config", "production.json"),
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
		config.MustGetFloat("bool")
	})

	closer.Close()
}

func TestStaticConfigHasOwnState(t *testing.T) {
	dict := map[string]interface{}{
		"a.b.c": "v",
		"bool":  true,
		"int":   int64(1),
		"float": float64(1.0),
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
