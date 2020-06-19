package bazhandler_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/baz/baz"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
)

func TestBazCall(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	// fixtures
	reqHeaders := map[string]string{
		"x-token":               "token",
		"x-uuid":                "uuid",
		"x-nil-response-header": "false",
	}
	args := &baz.SimpleService_Call_Args{
		Arg: &baz.BazRequest{
			B1: true,
			S2: "hello",
			I3: 42,
		},
	}
	expectedReqHeaders := map[string]string{
		"x-uuid":                "uuid",
		"x-nil-response-header": "false",
	}
	expectedResHeaders := map[string]string{
		"some-res-header": "something",
	}

	ctx := context.Background()
	var result baz.SimpleService_Call_Result
	ms.MockClients().Baz.EXPECT().Call(gomock.Any(), expectedReqHeaders, gomock.Any()).
		Return(map[string]string{"some-res-header": "something"}, nil)

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result,
	)
	if !assert.NoError(t, err, "got tchannel error") {
		return
	}
	assert.True(t, success)
	assert.Equal(t, expectedResHeaders, resHeaders)

	reqHeaders = map[string]string{
		"x-token":               "token",
		"x-uuid":                "uuid",
		"x-nil-response-header": "true",
	}
	expectedReqHeaders = map[string]string{
		"x-uuid":                "uuid",
		"x-nil-response-header": "true",
	}
	ms.MockClients().Baz.EXPECT().Call(gomock.Any(), expectedReqHeaders, gomock.Any()).
		Return(map[string]string{"some-res-header": "something"}, nil)

	success, _, err = ms.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result,
	)
	if !assert.Error(t, err, "got tchannel error") {
		return
	}
	assert.False(t, success)
}
