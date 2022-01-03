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

package zanzibar

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"
)

// HTTPServer like, http.Server but improved management of listening and serving
// Allows you to listen on port 0, query for real OS port and then serve requests
type HTTPServer struct {
	*http.Server
	Logger *zap.Logger

	listeningSocket net.Listener
	closing         bool

	RealPort int32
	RealIP   string
	RealAddr string
}

// JustListen will only listen on port and query real addr
func (server *HTTPServer) JustListen() (net.Listener, error) {
	addr := server.Addr
	if addr == "" {
		/* coverage ignore next line */
		addr = ":http"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	server.listeningSocket = ln

	realAddr := ln.Addr().(*net.TCPAddr)
	server.RealPort = int32(realAddr.Port)
	server.RealIP = realAddr.IP.String()
	server.RealAddr = realAddr.IP.String() + ":" +
		strconv.Itoa(int(server.RealPort))

	return ln, nil
}

// JustServe will serve all incoming requests
func (server *HTTPServer) JustServe(waitGroup *sync.WaitGroup) {
	ln := server.listeningSocket.(*net.TCPListener)

	err := server.Serve(tcpKeepAliveListener{ln})
	if err != nil && !server.closing {
		/* coverage ignore next line */
		server.Logger.Error("Error http serving", zap.Error(err))
	}

	waitGroup.Done()
}

// Close the listening socket
func (server *HTTPServer) Close() {
	server.closing = true
	if server.listeningSocket == nil {
		/* coverage ignore next line */
		return
	}

	err := server.listeningSocket.Close()
	if err != nil {
		/* coverage ignore next line */
		server.Logger.Error("Error closing listening socket", zap.Error(err))
	}
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	_ = tc.SetKeepAlive(true)
	_ = tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
