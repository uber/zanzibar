// Copyright (c) 2023 Uber Technologies, Inc.
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

package zanzibar_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	zanzibar "github.com/uber/zanzibar/runtime"
	"go.uber.org/zap"
)

func TestErrorBuilder(t *testing.T) {
	eb := zanzibar.NewErrorBuilder("endpoint", "foo")
	err := errors.New("test error")
	zerr := eb.Error(err, zanzibar.TChannelError)

	assert.Equal(t, "endpoint::foo", zerr.ErrorLocation())
	assert.Equal(t, "TChannelError", zerr.ErrorType().String())
	assert.True(t, errors.Is(zerr, err))
}

func TestErrorBuilderLogFields(t *testing.T) {
	eb := zanzibar.NewErrorBuilder("client", "bar")
	testErr := errors.New("test error")
	table := []struct {
		err             error
		wantErrLocation string
		wantErrType     string
	}{
		{
			err:             eb.Error(testErr, zanzibar.ClientException),
			wantErrLocation: "client::bar",
			wantErrType:     "ClientException",
		},
		{
			err:             testErr,
			wantErrLocation: "~client::bar",
			wantErrType:     "ErrorType(0)",
		},
	}
	for i, tt := range table {
		t.Run("test"+strconv.Itoa(i), func(t *testing.T) {
			logFieldErrLocation := eb.LogFieldErrorLocation(tt.err)
			logFieldErrType := eb.LogFieldErrorType(tt.err)

			assert.Equal(t, zap.String("errorLocation", tt.wantErrLocation), logFieldErrLocation)
			assert.Equal(t, zap.String("errorType", tt.wantErrType), logFieldErrType)
		})
	}
}
