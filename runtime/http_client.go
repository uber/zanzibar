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

package zanzibar

import "net/http"
import (
	"net"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

// HTTPClient defines a http client.
type HTTPClient struct {
	gateway *Gateway

	Client  *http.Client
	Logger  *zap.Logger
	Scope   *tally.Scope
	BaseURL string
}

// NewHTTPClient will allocate a http client.
func NewHTTPClient(
	gateway *Gateway, baseURL string,
) *HTTPClient {
	scope := gateway.MetricScope.SubScope("") // TODO: get http client name
	timer := scope.Timer("dial")
	success := scope.Counter("dial.success")
	error := scope.Counter("dial.error")

	return &HTTPClient{
		gateway: gateway,
		Logger:  gateway.Logger,
		Client: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:   false,
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 500,
				Dial:                getInstrumentedDial(timer, success, error),
			},
		},
		BaseURL: baseURL,
	}
}

func getInstrumentedDial(timer tally.Timer, successCtr, errorCtr tally.Counter) func(string, string) (net.Conn, error) {
	return func(n, a string) (conn net.Conn, err error) {
		stat := instrument(timer, successCtr, errorCtr)
		defer func() {
			stat(err)
		}()

		return net.Dial(n, a)
	}
}

func instrument(timer tally.Timer, successCtr, errorCtr tally.Counter) func(error) {
	start := time.Now()

	return func(err error) {
		timer.Record(time.Since(start))
		if err == nil {
			successCtr.Inc(1)
		} else {
			errorCtr.Inc(1)
		}
	}
}
