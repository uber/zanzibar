package bounce_test

import (
	"context"
	"testing"
	"time"

	mock "github.com/uber/zanzibar/examples/selective-gateway/build/services/selective-gateway/mock-service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/zanzibar/examples/selective-gateway/build/gen-code/endpoints/bounce/bounce"
	"github.com/uber/zanzibar/examples/selective-gateway/build/proto-gen/clients/mirror"
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

	ms.MockClients().Echo.ExpectEchoEcho().Success()
	ms.MockClients().Mirror.EXPECT().MirrorMirror(gomock.Any(), &mirror.Request{Message: message}).
		Return(ctx, &mirror.Response{Message: message}, nil)
	ms.MockClients().Mirror.EXPECT().MirrorInternalMirror(gomock.Any(), &mirror.InternalRequest{Message: message}).
		Return(ctx, &mirror.InternalResponse{Message: message}, nil)

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "Bounce", "bounce", nil, args, &result, &runtime.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(2000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: runtime.DefaultBackOffTimeAcrossRetries,
		},
	)
	require.NoError(t, err, "got tchannel error")

	assert.True(t, success)
	assert.Nil(t, resHeaders)
	assert.Equal(t, message, *result.Success)
}
