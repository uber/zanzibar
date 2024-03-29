package echo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/echo/echo"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/echo-gateway/mock-service"
)

func TestEcho(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	message := "hello"
	args := &echo.Echo_Echo_Args{
		Msg: message,
	}

	ctx := context.Background()
	var result echo.Echo_Echo_Result

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "Echo", "echo", nil, args, &result)
	if !assert.NoError(t, err, "got tchannel error") {
		return
	}
	assert.True(t, success)
	assert.Nil(t, resHeaders)
	assert.Equal(t, message, *result.Success)
}
