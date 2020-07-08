package bounce_test

import (
	"context"
	"testing"

	mock "github.com/uber/zanzibar/examples/selective-gateway/build/services/selective-gateway/mock-service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/zanzibar/examples/selective-gateway/build/gen-code/clients/echo"
	"github.com/uber/zanzibar/examples/selective-gateway/build/gen-code/clients/mirror"
	"github.com/uber/zanzibar/examples/selective-gateway/build/gen-code/endpoints/bounce/bounce"
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

	ms.MockClients().Echo.EXPECT().Echo(gomock.Any(), &echo.Request{Message: message}).
		Return(&echo.Response{Message: message}, nil)
	ms.MockClients().Mirror.EXPECT().Mirror(gomock.Any(), &mirror.Request{Message: message}).
		Return(&mirror.Response{Message: message}, nil)

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "Bounce", "bounce", nil, args, &result,
	)
	require.NoError(t, err, "got tchannel error")

	assert.True(t, success)
	assert.Nil(t, resHeaders)
	assert.Equal(t, message, *result.Success)
}
