package panichandler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/uber/zanzibar/v1/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/baz/baz"
	ms "github.com/uber/zanzibar/v1/examples/example-gateway/build/services/example-gateway/mock-service"
)

func TestPanicCall(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	// fixtures
	reqHeaders := map[string]string{
		"x-token":               "token",
		"x-uuid":                "uuid",
		"x-nil-response-header": "false",
	}
	args := &baz.SimpleService_AnotherCall_Args{
		Arg: &baz.BazRequest{
			B1: true,
			S2: "hello",
			I3: 42,
		},
	}

	ctx := context.Background()
	var result baz.SimpleService_AnotherCall_Result
	_, _, err := ms.MakeTChannelRequest(
		ctx, "SimpleService", "AnotherCall", reqHeaders, args, &result,
	)

	assert.Error(t, err, "got tchannel error")
}
