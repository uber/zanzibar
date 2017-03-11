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

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/kardianos/osext"
	"github.com/uber-go/zap"
)

var logger = zap.New(zap.NewJSONEncoder())

func spawnBenchServer(dirName string) *exec.Cmd {
	benchServerPath := path.Join(dirName, "..", "benchserver", "benchserver")

	var benchServerCmd *exec.Cmd
	if runtime.GOOS == "linux" {
		benchServerCmd = exec.Command("taskset", "-c", "1,2", benchServerPath)
	} else {
		benchServerCmd = exec.Command(benchServerPath)
	}

	benchServerCmd.Stdout = os.Stdout
	benchServerCmd.Stderr = os.Stderr

	err := benchServerCmd.Start()
	if err != nil {
		panic(err)
	}

	return benchServerCmd
}

func writeConfigToFile(config map[string]interface{}) (string, error) {
	tempConfigDir, err := ioutil.TempDir("", "zanzibar-bench-config-json")
	if err != nil {
		return "", err
	}

	jsonFile := path.Join(tempConfigDir, "production.json")
	configBytes, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(jsonFile, configBytes, os.ModePerm)
	if err != nil {
		return "", err
	}

	return tempConfigDir, nil
}

func spawnGateway(dirName string) *exec.Cmd {
	logTempDir, err := ioutil.TempDir("", "zanzibar-log-file")
	if err != nil {
		panic(err)
	}

	config := map[string]interface{}{
		"ip":                     "127.0.0.1",
		"port":                   8093,
		"tchannel.serviceName":   "bench-gateway",
		"tchannel.processName":   "bench-gateway",
		"metrics.m3.hostPort":    "127.0.0.1:8053",
		"metrics.tally.service":  "bench-gateway",
		"logger.fileName":        path.Join(logTempDir, "zanzibar.log"),
		"clients.contacts.port":  8092,
		"clients.googleNow.port": 8092,
		"clients.contacts.ip":    "127.0.0.1",
	}

	tempConfigDir, err := writeConfigToFile(config)
	if err != nil {
		panic(err)
	}
	uberConfigDir := tempConfigDir

	mainGatewayPath := path.Join(
		dirName, "..", "..", "examples",
		"example-gateway", "build", "example-gateway",
	)

	var gatewayCmd *exec.Cmd
	if runtime.GOOS == "linux" {
		gatewayCmd = exec.Command("taskset", "-c", "0,3", mainGatewayPath)
	} else {
		gatewayCmd = exec.Command(mainGatewayPath)
	}

	gatewayCmd.Env = append(os.Environ(), "CONFIG_DIR="+uberConfigDir)
	gatewayCmd.Env = append(gatewayCmd.Env, "ENVIRONMENT=production")
	gatewayCmd.Stderr = os.Stderr
	gatewayCmd.Stdout = os.Stdout

	err = gatewayCmd.Start()
	if err != nil {
		panic(err)
	}

	logger.Info("started main gateway",
		zap.String("baseYamlFile", tempConfigDir),
	)

	return gatewayCmd
}

func main() {
	execFile, err := osext.Executable()
	if err != nil {
		panic(err)
	}

	dirName := path.Dir(execFile)
	defaultBenchProgram := path.Join(dirName, "..", "contacts_1KB.lua")

	loadtest := flag.Bool("loadtest", false, "turn on wrk load testing")
	luaScript := flag.String("script", defaultBenchProgram, "wrk lua script to run")
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var loadTestScript string
	if path.IsAbs(*luaScript) {
		loadTestScript = *luaScript
	} else {
		loadTestScript = path.Join(cwd, *luaScript)
	}

	benchServerCmd := spawnBenchServer(dirName)
	gatewayCmd := spawnGateway(dirName)

	if *loadtest {
		spawnWrkLoadTest(loadTestScript)
		err = gatewayCmd.Process.Kill()
		if err != nil {
			panic(err)
		}
		err = benchServerCmd.Process.Kill()
		if err != nil {
			panic(err)
		}
	} else {
		err = gatewayCmd.Wait()
		if err != nil {
			panic(err)
		}
		err = benchServerCmd.Wait()
		if err != nil {
			panic(err)
		}
	}
}

func spawnWrkLoadTest(loadTestScript string) {
	loadTestContent, err := ioutil.ReadFile(loadTestScript)
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(bytes.NewReader(loadTestContent))
	line, _ := reader.ReadString('\n')

	// First line is the wrk Command
	line = line[3 : len(line)-1]

	segments := strings.Split(line, " ")

	time.Sleep(2 * time.Second)
	logger.Info("spawning wrk child process\n")

	wrkCmd := exec.Command("wrk", segments[1:]...)
	wrkCmd.Stdout = os.Stdout
	wrkCmd.Stderr = os.Stderr

	err = wrkCmd.Start()
	if err != nil {
		panic(err)
	}

	err = wrkCmd.Wait()
	if err != nil {
		panic(err)
	}
}
