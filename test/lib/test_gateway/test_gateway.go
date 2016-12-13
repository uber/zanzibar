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

package testGateway

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"

	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally/m3"
	config "github.com/uber/zanzibar/examples/example-gateway/config"
	"github.com/uber/zanzibar/test/lib/test_m3_server"
)

var realAddrRegex = regexp.MustCompile(
	`"realAddr":"([0-9\.\:]+)"`,
)

// TestGateway for testing
type TestGateway struct {
	cmd            *exec.Cmd
	binaryFileInfo *testBinaryInfo
	jsonLines      []string
	test           *testing.T
	opts           *Options
	httpClient     *http.Client
	m3Server       *testM3Server.FakeM3Server

	M3Service        *testM3Server.FakeM3Service
	MetricsWaitGroup sync.WaitGroup
	RealAddr         string
	RealHost         string
	RealPort         int
}

func getProjectDir() string {
	goPath := os.Getenv("GOPATH")
	return path.Join(goPath, "src", "github.com", "uber", "zanzibar")
}

// MalformedStdoutError is used when the child process has unexpected stdout
type MalformedStdoutError struct {
	Type       string
	StdoutLine string
	Message    string
}

func (err *MalformedStdoutError) Error() string {
	return err.Message
}

func readAddrFromStdout(testGateway *TestGateway, reader *bufio.Reader) error {
	line1, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	if line1[0] == '{' {
		testGateway.jsonLines = append(testGateway.jsonLines, line1)
	}

	_, err = os.Stdout.WriteString(line1)
	if err != nil {
		return err
	}

	m := realAddrRegex.FindStringSubmatch(line1)
	if m == nil {
		return &MalformedStdoutError{
			Type:       "malformed.stdout",
			StdoutLine: line1,
			Message: fmt.Sprintf(
				"Could not find RealAddr in server stdout: %s",
				line1,
			),
		}
	}

	testGateway.RealAddr = m[1]
	indexOfSep := strings.LastIndex(testGateway.RealAddr, ":")
	if indexOfSep != -1 {
		host := testGateway.RealAddr[0:indexOfSep]
		port := testGateway.RealAddr[indexOfSep+1:]
		portNum, err := strconv.Atoi(port)

		testGateway.RealHost = host
		if err != nil {
			testGateway.RealPort = -1
		} else {
			testGateway.RealPort = portNum
		}
	}

	return nil
}

// Options used to create TestGateway
type Options struct {
	LogWhitelist map[string]bool
	CountMetrics bool
}

// CreateGateway bootstrap gateway for testing
func CreateGateway(
	t *testing.T, config *config.Config, opts *Options,
) (*TestGateway, error) {
	config.IP = "127.0.0.1"

	countMetrics := false
	if opts != nil {
		countMetrics = opts.CountMetrics
	}

	testGateway := &TestGateway{
		test: t, opts: opts,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
			},
		},
	}
	testGateway.m3Server = testM3Server.NewFakeM3Server(
		t, &testGateway.MetricsWaitGroup,
		false, countMetrics, metrics.Compact,
	)
	testGateway.M3Service = testGateway.m3Server.Service
	go testGateway.m3Server.Serve()

	config.Metrics.M3.HostPort = testGateway.m3Server.Addr
	config.Metrics.Tally.Service = "test-example-gateway"
	config.Metrics.M3.FlushInterval = 10 * time.Millisecond
	config.Metrics.Tally.FlushInterval = 10 * time.Millisecond

	info, err := createTestBinaryFile(config)
	if err != nil {
		return nil, err
	}

	testGateway.binaryFileInfo = info

	args := []string{
		"-c", "0", testGateway.binaryFileInfo.binaryFile,
	}

	if os.Getenv("COVER_ON") == "1" {
		args = append(args,
			"-test.coverprofile", info.coverProfileFile,
		)
	}

	if runtime.GOOS == "linux" {
		testGateway.cmd = exec.Command("taskset", args...)
	} else {
		testGateway.cmd = exec.Command(args[2], args[3:]...)
	}
	tempConfigDir, err := writeConfigToFile(config)
	if err != nil {
		testGateway.Close()
		return nil, err
	}
	testGateway.cmd.Env = append(
		[]string{
			"UBER_CONFIG_DIR=" + tempConfigDir,
			"GATEWAY_RUN_CHILD_PROCESS_TEST=1",
		},
		os.Environ()...,
	)
	testGateway.cmd.Stderr = os.Stderr

	cmdStdout, err := testGateway.cmd.StdoutPipe()
	if err != nil {
		testGateway.Close()
		return nil, err
	}

	err = testGateway.cmd.Start()
	if err != nil {
		testGateway.Close()
		return nil, err
	}

	reader := bufio.NewReader(cmdStdout)

	err = readAddrFromStdout(testGateway, reader)
	if err != nil {
		testGateway.Close()
		return nil, err
	}

	go testGateway.copyToStdout(cmdStdout)

	return testGateway, nil
}

func (gateway *TestGateway) copyToStdout(src io.Reader) {
	reader := bufio.NewReader(src)

	for true {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if line == "PASS\n" {
			continue
		} else if strings.Index(line, "coverage:") == 0 {
			continue
		}

		if line[0] == '{' {
			gateway.jsonLines = append(gateway.jsonLines, line)
		}

		_, err = os.Stdout.WriteString(line)
		if err != nil {
			// TODO: betterer...
			panic(err)
		}
	}
}

// MakeRequest helper
func (gateway *TestGateway) MakeRequest(
	method string, url string, body io.Reader,
) (*http.Response, error) {
	client := gateway.httpClient

	fullURL := "http://" + gateway.RealAddr + url

	req, err := http.NewRequest(method, fullURL, body)

	if err != nil {
		return nil, err
	}

	return client.Do(req)
}

// Close test gateway
func (gateway *TestGateway) Close() {
	if gateway.cmd != nil {
		err := syscall.Kill(gateway.cmd.Process.Pid, syscall.SIGUSR2)
		if err != nil {
			panic(err)
		}

		_ = gateway.cmd.Wait()
	}

	if gateway.binaryFileInfo != nil {
		gateway.binaryFileInfo.Cleanup()
	}

	if gateway.m3Server != nil {
		_ = gateway.m3Server.Close()
	}

	// Sanity verify jsonLines
	for _, line := range gateway.jsonLines {
		lineStruct := map[string]interface{}{}
		jsonErr := json.Unmarshal([]byte(line), &lineStruct)
		if !assert.NoError(gateway.test, jsonErr, "logs must be json") {
			return
		}

		level := lineStruct["level"].(string)
		if level != "error" {
			continue
		}

		msg := lineStruct["msg"].(string)
		if gateway.opts == nil || !gateway.opts.LogWhitelist[msg] {
			assert.Fail(gateway.test,
				"Got unexpected error log from example-gateway:", line,
			)
		}
	}
}
