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
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var realAddrRegex = regexp.MustCompile(
	`"realAddr":"([0-9\.\:]+)"`,
)

// MalformedStdoutError is used when the child process has unexpected stdout
type MalformedStdoutError struct {
	Type       string
	StdoutLine string
	Message    string
}

func (err *MalformedStdoutError) Error() string {
	return err.Message
}

func (gateway *ChildProcessGateway) createAndSpawnChild(
	config map[string]interface{},
) error {
	info, err := createTestBinaryFile(config)
	if err != nil {
		return errors.Wrap(err, "Could not create test binary file: ")
	}

	gateway.binaryFileInfo = info

	args := []string{
		gateway.binaryFileInfo.binaryFile,
	}

	if os.Getenv("COVER_ON") == "1" {
		args = append(args,
			"-test.coverprofile", info.coverProfileFile,
		)
	}

	gateway.cmd = exec.Command(args[0], args[1:]...)
	tempConfigDir, err := writeConfigToFile(config)
	if err != nil {
		gateway.Close()
		return errors.Wrap(err, "Could not exec test command")
	}
	gateway.cmd.Env = append(
		[]string{
			"CONFIG_DIR=" + tempConfigDir,
			"GATEWAY_RUN_CHILD_PROCESS_TEST=1",
		},
		os.Environ()...,
	)
	gateway.cmd.Stderr = os.Stderr

	cmdStdout, err := gateway.cmd.StdoutPipe()
	if err != nil {
		gateway.Close()
		return errors.Wrap(err, "Could not create stdout pipe")
	}

	err = gateway.cmd.Start()
	if err != nil {
		gateway.Close()
		return errors.Wrap(err, "Could not start test gateway")
	}

	reader := bufio.NewReader(cmdStdout)

	err = readAddrFromStdout(gateway, reader)
	if err != nil {
		gateway.Close()
		return errors.Wrap(err, "could not read addr from stdout")
	}

	go gateway.copyToStdout(cmdStdout)
	return nil
}

func readAddrFromStdout(testGateway *ChildProcessGateway, reader *bufio.Reader) error {
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

func (gateway *ChildProcessGateway) copyToStdout(src io.Reader) {
	reader := bufio.NewReader(src)

	for line, err := reader.ReadString('\n'); err != nil; {
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
