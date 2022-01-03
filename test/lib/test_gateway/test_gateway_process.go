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
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

var realHTTPAddrRegex = regexp.MustCompile(
	`"realHTTPAddr":"([0-9\.\:]+)"`,
)
var realTChannelAddrRegex = regexp.MustCompile(
	`"realTChannelAddr":"([0-9\.\:]+)"`,
)

var infoIgnoreList = map[string]bool{
	"Outbound connection is active.":            true,
	"Channel.Close called.":                     true,
	"Connection.Close called.":                  true,
	"Connection state updated in Close.":        true,
	"Connection state updated during shutdown.": true,
	"Removed peer from root peer list.":         true,
	"Inbound connection is active.":             true,
	"Channel closed.":                           true,
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

func (gateway *ChildProcessGateway) createAndSpawnChild(
	mainFile string,
	defaultConfigFiles []string,
	testConfigOverrides map[string]interface{},
) error {
	info, err := createTestBinaryFile(mainFile)
	if err != nil {
		return errors.Wrap(err, "Could not create test binary file: ")
	}

	gateway.binaryFileInfo = info

	args := []string{
		gateway.binaryFileInfo.BinaryFile,
	}

	if os.Getenv("COVER_ON") == "1" {
		args = append(args,
			"-test.coverprofile", info.CoverProfileFile,
		)
	}

	tempConfigFile, err := writeConfigToFile(testConfigOverrides)
	if err != nil {
		gateway.Close()
		return errors.Wrap(err, "Could not exec test command")
	}

	configFiles := append(defaultConfigFiles, tempConfigFile)
	args = append(args, "-config", strings.Join(configFiles, ";"))
	gateway.cmd = exec.Command(args[0], args[1:]...)

	gateway.cmd.Env = append(
		[]string{
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

	go gateway.copyToStdout(reader)
	return nil
}

func addJSONLine(gateway *ChildProcessGateway, line string) {
	gateway.jsonLines = append(gateway.jsonLines, line)

	lineStruct := map[string]interface{}{}
	jsonErr := json.Unmarshal([]byte(line), &lineStruct)
	if jsonErr != nil {
		// do not decode msg
		return
	}

	msg := lineStruct["msg"].(string)

	msgLogs := gateway.logMessages[msg]
	if msgLogs == nil {
		msgLogs = []LogMessage{lineStruct}
	} else {
		msgLogs = append(msgLogs, lineStruct)
	}
	gateway.logMessages[msg] = msgLogs
}

func readAddrFromStdout(testGateway *ChildProcessGateway, reader *bufio.Reader) error {
	var msg string
	var httpFound, tchannelFound []string
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}

		if line[0] == '{' {
			addJSONLine(testGateway, line)
			msg += line
		}

		printJSONLine(line)

		if httpFound == nil {

			httpFound = realHTTPAddrRegex.FindStringSubmatch(line)
			if httpFound != nil {

				testGateway.RealHTTPAddr = httpFound[1]
				indexOfSep := strings.LastIndex(testGateway.RealHTTPAddr, ":")
				if indexOfSep != -1 {
					host := testGateway.RealHTTPAddr[0:indexOfSep]
					port := testGateway.RealHTTPAddr[indexOfSep+1:]
					portNum, err := strconv.Atoi(port)

					testGateway.RealHTTPHost = host
					if err != nil {
						testGateway.RealHTTPPort = -1
					} else {
						testGateway.RealHTTPPort = portNum
					}
				} else {
					httpFound = nil
				}
			}

		}

		if tchannelFound == nil {
			tchannelFound = realTChannelAddrRegex.FindStringSubmatch(line)

			if tchannelFound != nil {
				testGateway.RealTChannelAddr = tchannelFound[1]
				indexOfSep := strings.LastIndex(testGateway.RealTChannelAddr, ":")
				if indexOfSep != -1 {
					host := testGateway.RealTChannelAddr[0:indexOfSep]
					port := testGateway.RealTChannelAddr[indexOfSep+1:]
					portNum, err := strconv.Atoi(port)

					testGateway.RealTChannelHost = host
					if err != nil {
						testGateway.RealTChannelPort = -1
					} else {
						testGateway.RealTChannelPort = portNum
					}
				} else {
					tchannelFound = nil
				}
			}
		}
		if httpFound != nil && tchannelFound != nil {
			return nil
		}
	}

	if httpFound == nil || tchannelFound == nil {
		return &MalformedStdoutError{
			Type:       "malformed.stdout",
			StdoutLine: msg,
			Message:    "Could not find real http/tchannel address in server stdout",
		}
	}

	return nil
}

func (gateway *ChildProcessGateway) copyToStdout(reader *bufio.Reader) {
	for {
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
			addJSONLine(gateway, line)
		}

		printJSONLine(line)
	}
}

func printJSONLine(line string) {
	lineStruct := map[string]interface{}{}
	jsonErr := json.Unmarshal([]byte(line), &lineStruct)
	if jsonErr == nil {
		// Validate the whitelist if its valid json.

		msg := lineStruct["msg"].(string)

		// Do not print a line if its in the ignore list.
		if infoIgnoreList[msg] {
			return
		}
	}

	_, err := os.Stdout.WriteString(line)
	if err != nil {
		// TODO: betterer...
		panic(err)
	}
}
