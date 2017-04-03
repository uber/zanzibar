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

package testBackend

import (
	"net"
	"strconv"

	"github.com/uber-go/zap"
	"github.com/uber/tchannel-go"

	zt "github.com/uber/zanzibar/runtime/tchannel"
)

// TestTChannelBackend will pretend to be a http backend
type TestTChannelBackend struct {
	Channel  *tchannel.Channel
	Server   *zt.Server
	IP       string
	Port     int32
	RealPort int32
	RealAddr string
}

// BuildTChannelBackends returns a map of TChannel backends based on config
func BuildTChannelBackends(
	cfg map[string]interface{}, knownTChannelBackends []string,
) (map[string]*TestTChannelBackend, error) {
	n := len(knownTChannelBackends)
	result := make(map[string]*TestTChannelBackend, n)

	for i := 0; i < n; i++ {
		serviceName := knownTChannelBackends[i]
		backend, err := CreateTChannelBackend(0, serviceName)
		if err != nil {
			return nil, err
		}

		err = backend.Bootstrap()
		if err != nil {
			return nil, err
		}

		result[serviceName] = backend
		cfg["clients."+serviceName+".ip"] = "127.0.0.1"
		cfg["clients."+serviceName+".port"] = int64(backend.RealPort)
	}

	return result, nil
}

// Bootstrap creates a backend for testing
func (backend *TestTChannelBackend) Bootstrap() error {
	addr := backend.IP + ":" + strconv.Itoa(int(backend.Port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	realAddr := ln.Addr().(*net.TCPAddr)
	backend.RealPort = int32(realAddr.Port)
	backend.RealAddr = realAddr.IP.String() + ":" + strconv.Itoa(int(backend.RealPort))

	// tchannel serve does not block, connection handling is done in different goroutine
	err = backend.Channel.Serve(ln)
	return err
}

// Register registers tchannel server handler
func (backend *TestTChannelBackend) Register(server zt.TChanServer) {
	backend.Server.Register(server)
}

// Close closes the underlying channel
func (backend *TestTChannelBackend) Close() {
	backend.Channel.Close()
}

// CreateTChannelBackend creates a TChannel backend for testing
func CreateTChannelBackend(port int32, serviceName string) (*TestTChannelBackend, error) {
	backend := &TestTChannelBackend{
		IP:   "127.0.0.1",
		Port: port,
	}

	testLogger := zap.New(zap.NewJSONEncoder())

	tchannelOpts := &tchannel.ChannelOptions{
		Logger: tchannel.NullLogger,
	}

	channel, err := tchannel.NewChannel(serviceName, tchannelOpts)
	if err != nil {
		return nil, err
	}

	backend.Channel = channel
	backend.Server = zt.NewServer(channel, testLogger)

	return backend, nil
}
