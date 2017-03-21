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

package tchannel_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	tchan "github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/testutils"
	"github.com/uber/tchannel-go/thrift"
	"github.com/uber/zanzibar/runtime/tchannel"
	"github.com/uber/zanzibar/runtime/tchannel/gen-code/meta"

	"github.com/stretchr/testify/assert"
	"go.uber.org/thriftrw/ptr"
)

func TestDefaultHealth(t *testing.T) {
	withMetaSetup(t, func(ctx thrift.Context, c tchannel.TchanMeta, server *tchannel.Server) {
		ret, err := c.Health(ctx)
		if assert.NoError(t, err, "Health endpoint failed") {
			assert.True(t, ret.Ok, "Health status mismatch")
			assert.Nil(t, ret.Message, "Health message mismatch")
		}
	})
}

func TestVersionInfo(t *testing.T) {
	withMetaSetup(t, func(ctx thrift.Context, c tchannel.TchanMeta, server *tchannel.Server) {
		ret, err := c.VersionInfo(ctx)
		if assert.NoError(t, err, "VersionInfo endpoint failed") {
			expected := &meta.VersionInfo{
				Language:        "go",
				LanguageVersion: strings.TrimPrefix(runtime.Version(), "go"),
				Version:         tchan.VersionInfo,
			}
			assert.Equal(t, expected, ret, "Unexpected version info")
		}
	})
}

func customHealthEmpty(ctx thrift.Context) (bool, string) {
	return false, ""
}

func TestCustomHealthEmpty(t *testing.T) {
	withMetaSetup(t, func(ctx thrift.Context, c tchannel.TchanMeta, server *tchannel.Server) {
		server.RegisterHealthHandler(customHealthEmpty)
		ret, err := c.Health(ctx)
		if assert.NoError(t, err, "Health endpoint failed") {
			assert.False(t, ret.Ok, "Health status mismatch")
			assert.Nil(t, ret.Message, "Health message mismatch")
		}
	})
}

func customHealthNoEmpty(ctx thrift.Context) (bool, string) {
	return false, "from me"
}

func TestCustomHealthNoEmpty(t *testing.T) {
	withMetaSetup(t, func(ctx thrift.Context, c tchannel.TchanMeta, server *tchannel.Server) {
		server.RegisterHealthHandler(customHealthNoEmpty)
		ret, err := c.Health(ctx)
		if assert.NoError(t, err, "Health endpoint failed") {
			assert.False(t, ret.Ok, "Health status mismatch")
			assert.Equal(t, ret.Message, ptr.String("from me"), "Health message mismatch")
		}
	})
}

func withMetaSetup(t *testing.T, f func(ctx thrift.Context, c tchannel.TchanMeta, server *tchannel.Server)) {
	ctx, cancel := thrift.NewContext(time.Second * 10)
	defer cancel()

	// Start server
	tchan, server := setupMetaServer(t)
	defer tchan.Close()

	// Get client
	c := getMetaClient(t, tchan.PeerInfo().HostPort)
	f(ctx, c, server)
}

func setupMetaServer(t *testing.T) (*tchan.Channel, *tchannel.Server) {
	tchan := testutils.NewServer(t, testutils.NewOpts().SetServiceName("meta"))
	server := tchannel.NewServer(tchan)
	return tchan, server
}

func getMetaClient(t *testing.T, dst string) tchannel.TchanMeta {
	tchan := testutils.NewClient(t, nil)
	tchan.Peers().Add(dst)
	thriftClient := tchannel.NewClient(tchan, "meta", nil)
	return tchannel.NewTChanMetaClient(thriftClient)
}
