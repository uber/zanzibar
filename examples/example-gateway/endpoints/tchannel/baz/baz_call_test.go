package bazHandler_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/tchannel/baz/baz"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
)

func TestBazCall(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	// fixtures
	reqHeaders := map[string]string{
		"x-token": "token",
		"x-uuid":  "uuid",
	}
	args := &baz.SimpleService_Call_Args{
		Arg: &baz.BazRequest{
			B1: true,
			S2: "hello",
			I3: 42,
		},
	}
	expectedHeaders := map[string]string{
		"some-res-header": "something",
	}

	ctx := context.Background()
	var result baz.SimpleService_Call_Result
	ms.MockClientNodes().Baz.EXPECT().Call(gomock.Any(), reqHeaders, gomock.Any()).
		Return(map[string]string{"some-res-header": "something"}, nil)

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result,
	)
	if !assert.NoError(t, err, "got tchannel error") {
		return
	}
	assert.True(t, success)
	assert.Equal(t, expectedHeaders, resHeaders)
}
