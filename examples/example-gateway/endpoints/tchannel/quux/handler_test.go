package quuxhandler_test

import (
	"context"
	"testing"
	"time"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/quux/quux"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"

	// "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	runtime "github.com/uber/zanzibar/runtime"
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

	_, _, err := ms.MakeTChannelRequest(context.Background(), "SimpleService", "EchoString", nil, args, &result,
		&runtime.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: runtime.DefaultBackOffTimeAcrossRetries,
		})
	require.NoError(t, err, "got tchannel error")
	assert.Equal(t, "echo", *result.Success)
}
