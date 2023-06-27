package bazhandler_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/baz/baz"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
)

func TestBazEcho(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	message := "hello"
	args := &baz.SimpleService_Echo_Args{
		Msg: message,
	}

	ctx := context.Background()
	var result baz.SimpleService_Echo_Result

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "SimpleService", "Echo", nil, args, &result,
	)
	if !assert.NoError(t, err, "got tchannel error") {
		return
	}
	assert.True(t, success)
	assert.Nil(t, resHeaders)
	assert.Equal(t, message, *result.Success)
}
