package quuxhandler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/zanzibar/v2/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/quux/quux"
	ms "github.com/uber/zanzibar/v2/examples/example-gateway/build/services/example-gateway/mock-service"
)

func TestEchoStringFixture(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	ms.MockClients().Quux.ExpectEchoString().Success().MinTimes(1)

	var result quux.SimpleService_EchoString_Result
	args := &quux.SimpleService_EchoString_Args{
		Msg: "echo",
	}

	_, _, err := ms.MakeTChannelRequest(context.Background(), "SimpleService", "EchoString", nil, args, &result)
	require.NoError(t, err, "got tchannel error")
	assert.Equal(t, "echo", *result.Success)
}
