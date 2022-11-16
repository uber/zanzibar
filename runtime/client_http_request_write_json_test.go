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
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"github.com/uber/zanzibar/runtime/jsonwrapper"

	"go.uber.org/zap"
)

type writeJSONSuit struct {
	suite.Suite
	req             *ClientHTTPRequest
	expectedRawBody []byte
}

func (wjs *writeJSONSuit) SetupSuite() {
	client := NewHTTPClient(
		NewContextLogger(zap.NewNop()),
		tally.NewTestScope("", nil),
		jsonwrapper.NewDefaultJSONWrapper(),
		"foo",
		map[string]string{
			"bar": "foo::bar",
		},
		"test",
		nil,
		time.Microsecond*time.Duration(20),
	)
	wjs.req = NewClientHTTPRequest(context.TODO(), "foo", "bar", "foo::bar", client,
		&TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: DefaultBackOffTimeAcrossRetries,
		})
	wjs.expectedRawBody = []byte("{\"field\":\"hello\"}")

}

type myType struct{}

func (m *myType) MarshalJSON() ([]byte, error) {
	s := "{\"field\":\"hello\"}"
	return []byte(s), nil
}

func (wjs *writeJSONSuit) TestWriteJSONCustomMarshaler() {
	m := &myType{}
	err := wjs.req.WriteJSON("POST", "test", nil, m)
	assert.NoError(wjs.T(), err)
	assert.Equal(wjs.T(), wjs.expectedRawBody, wjs.req.rawBody)
}

type myTypeError struct {
	Field string
}

func (m *myTypeError) MarshalJSON() ([]byte, error) {
	return nil, errors.New("can not marshal")
}

func (wjs *writeJSONSuit) TestWriteJSONCustomMarshalerError() {
	m := &myTypeError{"hello"}
	err := wjs.req.WriteJSON("POST", "test", nil, m)
	assert.EqualError(wjs.T(), err, "Could not serialize foo.bar request json: json: error calling MarshalJSON for type *zanzibar.myTypeError: can not marshal")
}

type myTypeDefault struct {
	Field string `json:"field"`
}

func (wjs *writeJSONSuit) TestWriteJSONDefaultMarshaler() {
	m := &myTypeDefault{"hello"}
	err := wjs.req.WriteJSON("POST", "test", nil, m)
	assert.NoError(wjs.T(), err)
	assert.Equal(wjs.T(), wjs.expectedRawBody, wjs.req.rawBody)
}

func TestWriteJSONSuite(t *testing.T) {
	suite.Run(t, new(writeJSONSuit))
}
