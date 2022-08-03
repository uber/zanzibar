package bounce_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/echo"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/bounce/bounce"
	mock "github.com/uber/zanzibar/examples/example-gateway/build/services/echo-gateway/mock-service"
	runtime "github.com/uber/zanzibar/runtime"
)

func TestEcho(t *testing.T) {
	ms := mock.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	message := "hello"
	args := &bounce.Bounce_Bounce_Args{
		Msg: message,
	}

	ctx := context.Background()
	var result bounce.Bounce_Bounce_Result

	ms.MockClients().Echo.EXPECT().EchoEcho(gomock.Any(), &echo.Request{Message: message}).
		Return(ctx, &echo.Response{Message: message}, nil)

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "Bounce", "bounce",
		nil,
		args,
		&result,
		&runtime.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(4000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  0,
			BackOffTimeAcrossRetriesInMs: runtime.DefaultBackOffTimeAcrossRetries,
		},
	)
	require.NoError(t, err, "got tchannel error")

	assert.True(t, success)
	assert.Nil(t, resHeaders)
	assert.Equal(t, message, *result.Success)
}
