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

package echo_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"

	"github.com/uber/zanzibar/config"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/echo"
	echoclient "github.com/uber/zanzibar/examples/example-gateway/clients/echo"
	zanzibar "github.com/uber/zanzibar/runtime"
)

type echoServer struct{}

func (s *echoServer) Echo(ctx context.Context, r *echo.Request) (*echo.Response, error) {
	return &echo.Response{Message: r.Message}, nil
}

func TestEcho(t *testing.T) {
	sc := config.NewRuntimeConfigOrDie([]string{"../../../examples/example-gateway/config/test.yaml"}, nil)
	ip := sc.MustGetString("sidecarRouter.default.grpc.ip")
	port := sc.MustGetInt("sidecarRouter.default.grpc.port")
	address := fmt.Sprintf("%s:%d", ip, port)

	clientServiceNameMapping := make(map[string]string)
	sc.MustGetStruct("grpc.clientServiceNameMapping", &clientServiceNameMapping)

	listener, err := net.Listen("tcp", address)
	require.NoError(t, err)

	clientName := "echo"

	grpcTransport := grpc.NewTransport()
	inbound := grpcTransport.NewInbound(listener)
	outbound := grpcTransport.NewSingleOutbound(address)
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name:     clientName,
		Inbounds: yarpc.Inbounds{inbound},
		Outbounds: yarpc.Outbounds{
			clientName: transport.Outbounds{
				ServiceName: clientServiceNameMapping[clientName],
				Unary:       outbound,
			},
		},
	})
	handler := echo.BuildEchoYARPCProcedures(&echoServer{})
	dispatcher.Register(handler)

	client := echoclient.NewClient(&echoclient.Dependencies{
		Default: &zanzibar.DefaultDependencies{
			GRPCClientDispatcher: dispatcher,
			Config:               sc,
			Logger:               zap.NewNop(),
			ContextLogger:        zanzibar.NewContextLogger(zap.NewNop()),
			ContextMetrics:       zanzibar.NewContextMetrics(tally.NewTestScope("", nil)),
		},
	})

	err = dispatcher.Start()
	require.NoError(t, err)
	defer func() {
		err := dispatcher.Stop()
		require.NoError(t, err)
	}()

	message := "hello"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, res, err := client.EchoEcho(ctx, &echo.Request{Message: message})
	assert.NoError(t, err)
	assert.Equal(t, &echo.Response{Message: message}, res)
}
