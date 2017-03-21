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
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tchan "github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/testutils"
	"github.com/uber/tchannel-go/thrift"
	"github.com/uber/zanzibar/runtime/tchannel"
	gen "github.com/uber/zanzibar/test/runtime/tchannel/gen-code/baz"
	"github.com/uber/zanzibar/test/runtime/tchannel/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"golang.org/x/net/context"
)

// Generate the service mocks using go generate.
//go:generate mockery -dir ./gen-code/baz -name TChanSimpleService
//go:generate mockery -dir ./gen-code/baz -name TChanSecondService

type testArgs struct {
	server *tchannel.Server
	s1     *mocks.TChanSimpleService
	s2     *mocks.TChanSecondService
	c1     gen.TChanSimpleService
	c2     gen.TChanSecondService

	serverCh *tchan.Channel
	clientCh *tchan.Channel
}

func ctxArg() mock.AnythingOfTypeArgument {
	return mock.AnythingOfType("tchannel.headerCtx")
}

func TestThriftArgs(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		arg := &gen.Data{
			B1: true,
			S2: "str",
			I3: 102,
		}
		ret := &gen.Data{
			B1: false,
			S2: "return-str",
			I3: 105,
		}

		args.s1.On("Call", ctxArg(), arg).Return(ret, nil)
		got, err := args.c1.Call(ctx, arg)
		require.NoError(t, err)
		assert.Equal(t, ret, got)
	})
}

func TestRequest(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		args.s1.On("Simple", ctxArg()).Return(nil)
		require.NoError(t, args.c1.Simple(ctx))
	})
}

func TestRetryRequest(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		count := 0
		args.s1.On("Simple", ctxArg()).Return(tchan.ErrServerBusy).
			Run(func(args mock.Arguments) {
				count++
			})
		require.Error(t, args.c1.Simple(ctx), "Simple expected to fail")
		assert.Equal(t, 5, count, "Expected Simple to be retried 5 times")
	})
}

func TestRequestSubChannel(t *testing.T) {
	ctx, cancel := thrift.NewContext(time.Second)
	defer cancel()

	tch := testutils.NewServer(t, testutils.NewOpts().SetServiceName("svc1"))
	defer tch.Close()

	clientCh := testutils.NewClient(t, nil)
	defer clientCh.Close()
	clientCh.Peers().Add(tch.PeerInfo().HostPort)

	tests := []tchan.Registrar{tch, tch.GetSubChannel("svc2"), tch.GetSubChannel("svc3")}
	for _, ch := range tests {
		mockHandler := new(mocks.TChanSecondService)
		server := tchannel.NewServer(ch)
		server.Register(gen.NewTChanSecondServiceServer(mockHandler))

		client := tchannel.NewClient(clientCh, ch.ServiceName(), nil)
		secondClient := gen.NewTChanSecondServiceClient(client)

		echoArg := ch.ServiceName()
		echoRes := echoArg + "-echo"
		mockHandler.On("Echo", ctxArg(), echoArg).Return(echoRes, nil)
		res, err := secondClient.Echo(ctx, echoArg)
		assert.NoError(t, err, "Echo failed")
		assert.Equal(t, echoRes, res)
	}
}

func TestLargeRequest(t *testing.T) {
	arg := testutils.RandString(100000)
	res := strings.ToLower(arg)

	fmt.Println(len(arg))
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		args.s2.On("Echo", ctxArg(), arg).Return(res, nil)

		got, err := args.c2.Echo(ctx, arg)
		if assert.NoError(t, err, "Echo got error") {
			assert.Equal(t, res, got, "Echo got unexpected response")
		}
	})
}

func TestThriftError(t *testing.T) {
	thriftErr := &gen.SimpleErr{
		Message: "this is the error",
	}
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		args.s1.On("Simple", ctxArg()).Return(thriftErr)
		got := args.c1.Simple(ctx)
		require.Error(t, got)
		require.Equal(t, thriftErr, got)
	})
}

func TestThriftUnknownError(t *testing.T) {
	thriftErr := &gen.NewErr{
		Message: "new error",
	}

	withSetup(t, func(ctx thrift.Context, args testArgs) {
		// When "Simple" is called, actually call a separate similar looking method
		// SimpleFuture which has a new exception that the client side of Simple
		// does not know how to handle.
		args.s1.On("SimpleFuture", ctxArg()).Return(thriftErr)
		tClient := tchannel.NewClient(args.clientCh, args.serverCh.ServiceName(), nil)
		rewriteMethodClient := rewriteMethodClient{tClient, "SimpleFuture"}
		simpleClient := gen.NewTChanSimpleServiceClient(rewriteMethodClient)

		got := simpleClient.Simple(ctx)
		require.Error(t, got)
		assert.Contains(t, got.Error(), "no result or unknown exception")
	})
}

func TestThriftNilErr(t *testing.T) {
	var thriftErr *gen.SimpleErr
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		args.s1.On("Simple", ctxArg()).Return(thriftErr)
		got := args.c1.Simple(ctx)
		require.Error(t, got)
		require.Contains(t, got.Error(), "non-nil error type")
		require.Contains(t, got.Error(), "nil value")
	})
}

func TestThriftDecodeEmptyFrameServer(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		args.s1.On("Simple", ctxArg()).Return(nil)

		call, err := args.clientCh.BeginCall(ctx, args.serverCh.PeerInfo().HostPort, args.serverCh.ServiceName(), "SimpleService::Simple", nil)
		require.NoError(t, err, "Failed to BeginCall")

		withWriter(t, call.Arg2Writer, func(w tchan.ArgWriter) error {
			if err := thrift.WriteHeaders(w, nil); err != nil {
				return err
			}

			return w.Flush()
		})

		withWriter(t, call.Arg3Writer, func(w tchan.ArgWriter) error {
			if err := tchannel.WriteStruct(w, &gen.SimpleService_Simple_Args{}); err != nil {
				return err
			}

			return w.Flush()
		})

		response := call.Response()
		withReader(t, response.Arg2Reader, func(r tchan.ArgReader) error {
			_, err := thrift.ReadHeaders(r)
			return err
		})

		var res gen.SimpleService_Simple_Result
		withReader(t, response.Arg3Reader, func(r tchan.ArgReader) error {
			return tchannel.ReadStruct(r, &res)
		})

		assert.Nil(t, res.SimpleErr, "Expected no error")
	})
}

func TestThriftDecodeEmptyFrameClient(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		handler := func(ctx context.Context, call *tchan.InboundCall) {
			withReader(t, call.Arg2Reader, func(r tchan.ArgReader) error {
				_, err := thrift.ReadHeaders(r)
				return err
			})

			withReader(t, call.Arg3Reader, func(r tchan.ArgReader) error {
				req := &gen.SimpleService_Simple_Args{}
				return tchannel.ReadStruct(r, req)
			})

			response := call.Response()
			withWriter(t, response.Arg2Writer, func(w tchan.ArgWriter) error {
				if err := thrift.WriteHeaders(w, nil); err != nil {
					return err
				}

				return w.Flush()
			})

			withWriter(t, response.Arg3Writer, func(w tchan.ArgWriter) error {
				if err := tchannel.WriteStruct(w, &gen.SimpleService_Simple_Result{}); err != nil {
					return err
				}

				return w.Flush()
			})
		}

		args.serverCh.Register(tchan.HandlerFunc(handler), "SimpleService::Simple")
		require.NoError(t, args.c1.Simple(ctx))
	})
}

func TestUnknownError(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		args.s1.On("Simple", ctxArg()).Return(errors.New("unexpected err"))
		got := args.c1.Simple(ctx)
		require.Error(t, got)
		require.Equal(t, tchan.NewSystemError(tchan.ErrCodeUnexpected, "unexpected err"), got)
	})
}

func TestMultiple(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		args.s1.On("Simple", ctxArg()).Return(nil)
		args.s2.On("Echo", ctxArg(), "test1").Return("test2", nil)

		require.NoError(t, args.c1.Simple(ctx))
		res, err := args.c2.Echo(ctx, "test1")
		require.NoError(t, err)
		require.Equal(t, "test2", res)
	})
}

func TestHeaders(t *testing.T) {
	reqHeaders := map[string]string{"header1": "value1", "header2": "value2"}
	respHeaders := map[string]string{"resp1": "value1-resp", "resp2": "value2-resp"}

	withSetup(t, func(ctx thrift.Context, args testArgs) {
		args.s1.On("Simple", ctxArg()).Return(nil).Run(func(args mock.Arguments) {
			ctx := args.Get(0).(thrift.Context)
			assert.Equal(t, reqHeaders, ctx.Headers(), "request headers mismatch")
			ctx.SetResponseHeaders(respHeaders)
		})

		ctx = thrift.WithHeaders(ctx, reqHeaders)
		require.NoError(t, args.c1.Simple(ctx))
		assert.Equal(t, respHeaders, ctx.ResponseHeaders(), "response headers mismatch")
	})
}

func TestClientHostPort(t *testing.T) {
	ctx, cancel := thrift.NewContext(time.Second)
	defer cancel()

	s1ch := testutils.NewServer(t, nil)
	s2ch := testutils.NewServer(t, nil)
	defer s1ch.Close()
	defer s2ch.Close()

	s1ch.Peers().Add(s2ch.PeerInfo().HostPort)
	s2ch.Peers().Add(s1ch.PeerInfo().HostPort)

	mock1, mock2 := new(mocks.TChanSecondService), new(mocks.TChanSecondService)
	tchannel.NewServer(s1ch).Register(gen.NewTChanSecondServiceServer(mock1))
	tchannel.NewServer(s2ch).Register(gen.NewTChanSecondServiceServer(mock2))

	// When we call using a normal client, it can only call the other server (only peer).
	c1 := gen.NewTChanSecondServiceClient(tchannel.NewClient(s1ch, s2ch.PeerInfo().ServiceName, nil))
	mock2.On("Echo", ctxArg(), "call1").Return("call1", nil)
	res, err := c1.Echo(ctx, "call1")
	assert.NoError(t, err, "call1 failed")
	assert.Equal(t, "call1", res)

	// When we call using a client that specifies host:port, it should call that server.
	c2 := gen.NewTChanSecondServiceClient(tchannel.NewClient(s1ch, s1ch.PeerInfo().ServiceName, &tchannel.ClientOptions{
		HostPort: s1ch.PeerInfo().HostPort,
	}))
	mock1.On("Echo", ctxArg(), "call2").Return("call2", nil)
	res, err = c2.Echo(ctx, "call2")
	assert.NoError(t, err, "call2 failed")
	assert.Equal(t, "call2", res)
}

func TestRegisterPostResponseCB(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		var createdCtx thrift.Context
		ctxKey := "key"
		ctxValue := "value"

		args.server.SetContextFn(func(ctx context.Context, method string, headers map[string]string) thrift.Context {
			createdCtx = thrift.WithHeaders(context.WithValue(ctx, ctxKey, ctxValue), headers)
			return createdCtx
		})

		arg := &gen.Data{
			B1: true,
			S2: "str",
			I3: 102,
		}
		ret := &gen.Data{
			B1: false,
			S2: "return-str",
			I3: 105,
		}

		called := make(chan struct{})
		cb := func(reqCtx context.Context, method string, response tchannel.RWTStruct) {
			assert.Equal(t, "Call", method)
			assert.Equal(t, createdCtx, reqCtx)
			assert.Equal(t, ctxValue, reqCtx.Value(ctxKey))
			res, ok := response.(*gen.SimpleService_Call_Result)
			if assert.True(t, ok, "response type should be Result struct") {
				resp, err := gen.SimpleService_Call_Helper.UnwrapResponse(res)
				assert.NoError(t, err, "unwarp response failed")
				assert.Equal(t, ret, resp, "result should be returned value")
			}
			close(called)
		}
		args.server.Register(gen.NewTChanSimpleServiceServer(args.s1), tchannel.OptPostResponse(cb))

		args.s1.On("Call", ctxArg(), arg).Return(ret, nil)
		res, err := args.c1.Call(ctx, arg)
		require.NoError(t, err, "Call failed")
		assert.Equal(t, res, ret, "Call return value wrong")
		select {
		case <-time.After(time.Second):
			t.Error("post-response callback not called")
		case <-called:
		}
	})
}

func TestRegisterPostResponseCBCalledOnError(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		var createdCtx thrift.Context
		ctxKey := "key"
		ctxValue := "value"

		args.server.SetContextFn(func(ctx context.Context, method string, headers map[string]string) thrift.Context {
			createdCtx = thrift.WithHeaders(context.WithValue(ctx, ctxKey, ctxValue), headers)
			return createdCtx
		})

		arg := &gen.Data{
			B1: true,
			S2: "str",
			I3: 102,
		}

		retErr := errors.New("unexpected error")

		called := make(chan struct{})
		cb := func(reqCtx context.Context, method string, response tchannel.RWTStruct) {
			assert.Equal(t, "Call", method)
			assert.Equal(t, createdCtx, reqCtx)
			assert.Equal(t, ctxValue, reqCtx.Value(ctxKey))
			assert.Nil(t, response)
			close(called)
		}
		args.server.Register(gen.NewTChanSimpleServiceServer(args.s1), tchannel.OptPostResponse(cb))

		args.s1.On("Call", ctxArg(), arg).Return(nil, retErr)
		res, err := args.c1.Call(ctx, arg)
		require.Error(t, err, "Call succeeded instead of failed")
		require.Nil(t, res, "Call returned value and an error")
		sysErr, ok := err.(tchan.SystemError)
		require.True(t, ok, "Call return error not a system error")
		assert.Equal(t, tchan.ErrCodeUnexpected, sysErr.Code(), "Call return error value wrong")
		assert.Equal(t, retErr.Error(), sysErr.Message(), "Call return error value wrong")
		select {
		case <-time.After(time.Second):
			t.Error("post-response callback not called")
		case <-called:
		}
	})
}

func TestThriftTimeout(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		handler := make(chan struct{})

		args.s2.On("Echo", ctxArg(), "asd").Return("asd", nil).Run(func(args mock.Arguments) {
			time.Sleep(testutils.Timeout(15 * time.Millisecond))
			close(handler)
		})

		newCtx, cancel := thrift.NewContext(testutils.Timeout(10 * time.Millisecond))
		defer cancel()

		_, err := args.c2.Echo(newCtx, "asd")
		assert.Equal(t, err, tchan.ErrTimeout, "Expect call to time out")

		// Wait for the handler to return, otherwise the test ends before the Server gets an error.
		select {
		case <-handler:
		case <-time.After(time.Second):
			t.Error("Echo handler did not run")
		}
	})
}

func TestThriftContextFn(t *testing.T) {
	withSetup(t, func(ctx thrift.Context, args testArgs) {
		args.server.SetContextFn(func(ctx context.Context, method string, headers map[string]string) thrift.Context {
			return thrift.WithHeaders(ctx, map[string]string{"custom": "headers"})
		})

		args.s2.On("Echo", ctxArg(), "test").Return("test", nil).Run(func(args mock.Arguments) {
			ctx := args.Get(0).(thrift.Context)
			assert.Equal(t, "headers", ctx.Headers()["custom"], "Custom header is missing")
		})
		_, err := args.c2.Echo(ctx, "test")
		assert.NoError(t, err, "Echo failed")
	})
}

func withSetup(t *testing.T, f func(ctx thrift.Context, args testArgs)) {
	args := testArgs{
		s1: new(mocks.TChanSimpleService),
		s2: new(mocks.TChanSecondService),
	}

	newCtx, cancel := thrift.NewContext(time.Second)
	defer cancel()

	// Start server
	args.serverCh, args.server = setupServer(t, args.s1, args.s2)
	defer args.serverCh.Close()

	args.clientCh, args.c1, args.c2 = getClients(t, args.serverCh.PeerInfo(), args.serverCh.ServiceName(), args.clientCh)

	f(newCtx, args)

	args.s1.AssertExpectations(t)
	args.s2.AssertExpectations(t)
}

func setupServer(t *testing.T, h *mocks.TChanSimpleService, sh *mocks.TChanSecondService) (*tchan.Channel, *tchannel.Server) {
	ch := testutils.NewServer(t, nil)
	server := tchannel.NewServer(ch)
	server.Register(gen.NewTChanSimpleServiceServer(h))
	server.Register(gen.NewTChanSecondServiceServer(sh))
	return ch, server
}

func getClients(t *testing.T, serverInfo tchan.LocalPeerInfo, svcName string, clientCh *tchan.Channel) (*tchan.Channel, gen.TChanSimpleService, gen.TChanSecondService) {
	ch := testutils.NewClient(t, nil)

	ch.Peers().Add(serverInfo.HostPort)
	client := tchannel.NewClient(ch, svcName, nil)

	simpleClient := gen.NewTChanSimpleServiceClient(client)
	secondClient := gen.NewTChanSecondServiceClient(client)
	return ch, simpleClient, secondClient
}

func withReader(t *testing.T, readerFn func() (tchan.ArgReader, error), f func(r tchan.ArgReader) error) {
	reader, err := readerFn()
	require.NoError(t, err, "Failed to get reader")

	err = f(reader)
	require.NoError(t, err, "Failed to read contents")

	require.NoError(t, reader.Close(), "Failed to close reader")
}

func withWriter(t *testing.T, writerFn func() (tchan.ArgWriter, error), f func(w tchan.ArgWriter) error) {
	writer, err := writerFn()
	require.NoError(t, err, "Failed to get writer")

	err = f(writer)
	require.NoError(t, err, "Failed to write contents")

	require.NoError(t, writer.Close(), "Failed to close Writer")
}

type rewriteMethodClient struct {
	client    tchannel.TChanClient
	rewriteTo string
}

func (c rewriteMethodClient) Call(ctx thrift.Context, serviceName, methodName string, req, resp tchannel.RWTStruct) (success bool, err error) {
	return c.client.Call(ctx, serviceName, c.rewriteTo, req, resp)
}
