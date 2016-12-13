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
	"io/ioutil"
	"os/exec"
	"path"
	"strings"

	"os"

	"flag"

	"bytes"

	"time"

	osext "github.com/kardianos/osext"
	"github.com/uber-go/zap"
	config "github.com/uber/zanzibar/examples/example-gateway/config"
	yaml "gopkg.in/yaml.v2"
)

var logger = zap.New(zap.NewJSONEncoder())

func spawnBenchServer(dirName string) *exec.Cmd {
	benchServerPath := path.Join(dirName, "..", "benchserver", "benchserver")

	benchServerCmd := exec.Command("taskset", "-c", "1,2", benchServerPath)
	benchServerCmd.Stdout = os.Stdout
	benchServerCmd.Stderr = os.Stderr

	err := benchServerCmd.Start()
	if err != nil {
		panic(err)
	}

	return benchServerCmd
}

func writeConfig(gatewayConfig *config.Config) string {
	tempConfigDir, err := ioutil.TempDir("", "zanzibar-config-yaml")
	if err != nil {
		panic(err)
	}

	productionYamlFile := path.Join(tempConfigDir, "production.yaml")

	configBytes, err := yaml.Marshal(gatewayConfig)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(productionYamlFile, configBytes, os.ModePerm)
	if err != nil {
		panic(err)
	}

	return tempConfigDir
}

func spawnGateway(dirName string) *exec.Cmd {

	logTempDir, err := ioutil.TempDir("", "zanzibar-log-file")
	if err != nil {
		panic(err)
	}

	gatewayConfig := &config.Config{}
	gatewayConfig.Logger.FileName = path.Join(logTempDir, "zanzibar.log")
	gatewayConfig.Clients.Contacts.Port = 8092
	gatewayConfig.Clients.Contacts.IP = "127.0.0.1"
	gatewayConfig.IP = "127.0.0.1"
	gatewayConfig.Port = 8093
	gatewayConfig.Metrics.M3.HostPort = "127.0.0.1:8053"
	gatewayConfig.Metrics.Tally.Service = "bench-zanzibar"

	tempConfigDir := writeConfig(gatewayConfig)
	uberConfigDir := tempConfigDir

	mainGatewayPath := path.Join(
		dirName, "..", "..", "examples", "example-gateway", "example-gateway",
	)
	gatewayCmd := exec.Command("taskset", "-c", "0,3", mainGatewayPath)
	gatewayCmd.Env = append(os.Environ(), "UBER_CONFIG_DIR="+uberConfigDir)
	gatewayCmd.Env = append(gatewayCmd.Env, "UBER_ENVIRONMENT=production")
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
	loadTestScript := path.Join(dirName, "..", "contacts_1KB.lua")

	loadtest := flag.Bool("loadtest", false, "turn on wrk load testing")
	luascript := flag.String("script", loadTestScript, "wrk lua script to run")
	flag.Parse()

	benchServerCmd := spawnBenchServer(dirName)
	gatewayCmd := spawnGateway(dirName)

	if *loadtest {
		spawnWrkLoadTest(*luascript)
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
