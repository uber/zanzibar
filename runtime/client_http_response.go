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

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// ClientHTTPResponse is the struct managing the client response
// when making outbound http calls.
type ClientHTTPResponse struct {
	req              *ClientHTTPRequest
	finishTime       time.Time
	finished         bool
	rawResponse      *http.Response
	rawResponseBytes []byte

	StatusCode int
	Header     http.Header
}

// NewClientHTTPResponse allocates a client http response object
// to track http response.
func NewClientHTTPResponse(
	req *ClientHTTPRequest,
) *ClientHTTPResponse {
	res := &ClientHTTPResponse{
		req: req,
	}

	return res
}

func (res *ClientHTTPResponse) setRawHTTPResponse(httpRes *http.Response) {
	res.rawResponse = httpRes
	res.StatusCode = httpRes.StatusCode
	res.Header = httpRes.Header
}

// ReadAll reads bytes from response.
func (res *ClientHTTPResponse) ReadAll() ([]byte, error) {
	rawBody, err := ioutil.ReadAll(res.rawResponse.Body)

	cerr := res.rawResponse.Body.Close()
	if cerr != nil {
		/* coverage ignore next line */
		res.req.Logger.Error("Could not close client resp body",
			zap.String("error", err.Error()),
		)
	}

	if err != nil {
		res.req.Logger.Error("Could not ReadAll() client body",
			zap.String("error", err.Error()),
		)
		return nil, errors.Wrapf(
			err,
			"Could not read client(%s) response body",
			res.req.ClientID,
		)
	}

	res.rawResponseBytes = rawBody
	return rawBody, nil
}

// GetRawBody returns the body as byte array if it has been read.
func (res *ClientHTTPResponse) GetRawBody() []byte {
	return res.rawResponseBytes
}

// ReadAndUnmarshalBody will try to unmarshal non pointer value or fail
func (res *ClientHTTPResponse) ReadAndUnmarshalBody(v interface{}) error {
	rawBody, err := res.ReadAll()
	if err != nil {
		/* coverage ignore next line */
		return err
	}

	err = json.Unmarshal(rawBody, v)
	if err != nil {
		res.req.Logger.Warn("Could not parse client json",
			zap.String("error", err.Error()),
		)
		return errors.Wrapf(
			err,
			"Could not parse client %q json",
			res.req.ClientID,
		)
	}

	return nil
}

// CheckOKResponse checks if the status code is OK.
func (res *ClientHTTPResponse) CheckOKResponse(okResponses []int) {
	for _, okResponse := range okResponses {
		if res.rawResponse.StatusCode == okResponse {
			return
		}
	}

	res.req.Logger.Warn("Unknown response status code",
		zap.Int("status code", res.rawResponse.StatusCode),
	)
}

// finish()
