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

package testgateway

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	yaml "github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

var cachedBinaryFile *testBinaryInfo

type testBinaryInfo struct {
	BinaryFile       string
	CoverProfileFile string
	MainFile         string
}

func (info *testBinaryInfo) Cleanup() {
	if os.Getenv("COVER_ON") != "1" {
		return
	}

	randStr, err := makeRandStr()
	if err != nil {
		panic(err)
	}

	newCoverProfileFile := path.Join(
		path.Dir(info.CoverProfileFile),
		"cover-"+randStr+".out",
	)

	bytes, err := ioutil.ReadFile(info.CoverProfileFile)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(newCoverProfileFile, bytes, 0644)
	if err != nil {
		panic(err)
	}
}

func writeConfigToFile(config map[string]interface{}) (string, error) {
	tempConfigDir, err := ioutil.TempDir("", "example-gateway-config-yaml")
	if err != nil {
		return "", err
	}

	yamlFile := path.Join(tempConfigDir, "test.yaml")
	configBytes, err := yaml.Marshal(config)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(yamlFile, configBytes, os.ModePerm)
	if err != nil {
		return "", err
	}

	return yamlFile, nil
}

func makeRandStr() (string, error) {
	randBytes := make([]byte, 8)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}
	randStr := fmt.Sprintf("%x", randBytes)

	return randStr, nil
}

func getDirName() string {
	_, file, _, _ := runtime.Caller(0)

	return filepath.Dir(file)
}

func getZanzibarDirName() string {
	return filepath.Join(getDirName(), "..", "..", "..")
}

func tryLoadCachedBinaryTestInfo(mainPath string) {
	jsonFile := filepath.Join(
		getDirName(), "..", "..", ".cached_binary_test_info.json",
	)

	bytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return
	}

	jsonObj := &testBinaryInfo{}
	err = json.Unmarshal(bytes, jsonObj)
	if err != nil {
		return
	}

	if jsonObj.MainFile != mainPath {
		return
	}

	cachedBinaryFile = jsonObj
}

func tryWriteCachedBinaryTestInfo() {
	jsonFile := filepath.Join(
		getDirName(), "..", "..", ".cached_binary_test_info.json",
	)

	bytes, err := json.Marshal(cachedBinaryFile)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(jsonFile, bytes, 0644)
	if err != nil {
		return
	}
}

func createTestBinaryFile(mainPath string) (*testBinaryInfo, error) {
	if os.Getenv("ZANZIBAR_CACHE") == "1" && cachedBinaryFile == nil {
		// Try to load cachedBinaryFile from disk
		tryLoadCachedBinaryTestInfo(mainPath)
	}

	if cachedBinaryFile != nil && cachedBinaryFile.MainFile == mainPath {
		return cachedBinaryFile, nil
	}

	zanzibarRoot := getZanzibarDirName()

	mainTestPath := strings.Replace(mainPath, "main.go", "main_test.go", -1)

	randStr, err := makeRandStr()
	if err != nil {
		return nil, errors.Wrap(err, "could not make rand str...")
	}

	coverProfileFile := path.Join(
		zanzibarRoot, "coverage", "cover-"+randStr+".out",
	)

	tempConfigDir, err := ioutil.TempDir("", "example-gateway-test-binary")
	if err != nil {
		return nil, errors.Wrap(err, "could not create temp dir")
	}

	binaryFile := path.Join(tempConfigDir, "test-"+randStr+".test")

	novendorCmd := exec.Command("glide", "novendor")
	novendorCmd.Dir = zanzibarRoot
	outBytes, err := novendorCmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "could not run glide novendor command")
	}

	allFolders := strings.Split(string(outBytes), "\n")
	allPackages := []string{}
	for _, folder := range allFolders {
		if folder == "./test/..." ||
			folder == "./benchmarks/..." ||
			folder == "./scripts/..." ||
			folder == "" ||
			folder == "./workspace/..." {
			continue
		}

		allPackages = append(allPackages, path.Join("github.com/uber/zanzibar", folder))
	}

	allPackagesString := strings.Join(allPackages, ",")

	args := []string{"test", "-v", "-c", "-o", binaryFile}
	if os.Getenv("COVER_ON") == "1" {
		args = append(args,
			"-cover", "-coverprofile", coverProfileFile,
			"-coverpkg", allPackagesString)
	}

	args = append(args, mainTestPath, mainPath)

	testGenCmd := exec.Command("go", args...)

	testGenCmd.Stderr = os.Stderr
	testGenCmd.Stdout = os.Stdout
	err = testGenCmd.Run()
	if err != nil {
		return nil, errors.Wrap(err, "could not compile test program")
	}

	cachedBinaryFile = &testBinaryInfo{
		BinaryFile:       binaryFile,
		CoverProfileFile: coverProfileFile,
		MainFile:         mainPath,
	}
	if os.Getenv("ZANZIBAR_CACHE") == "1" {
		tryWriteCachedBinaryTestInfo()
	}
	return cachedBinaryFile, nil
}
