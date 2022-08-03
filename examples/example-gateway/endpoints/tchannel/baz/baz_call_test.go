package bazhandler_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/baz/baz"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
	runtime "github.com/uber/zanzibar/runtime"
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
	ms.MockClients().Baz.EXPECT().Call(gomock.Any(), expectedReqHeaders, gomock.Any(), gomock.Any()).
		Return(ctx, map[string]string{"some-res-header": "something"}, nil)

	success, resHeaders, err := ms.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result, &runtime.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: runtime.DefaultBackOffTimeAcrossRetries,
		},
	)
	dynamicRespHeaders := []string{
		"client.response.duration",
	}
	for _, dynamicValue := range dynamicRespHeaders {
		assert.Contains(t, resHeaders, dynamicValue)
		delete(resHeaders, dynamicValue)
	}

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
	ms.MockClients().Baz.EXPECT().Call(gomock.Any(), expectedReqHeaders, gomock.Any(), gomock.Any()).
		Return(ctx, map[string]string{"some-res-header": "something"}, nil)

	success, _, err = ms.MakeTChannelRequest(
		ctx, "SimpleService", "Call", reqHeaders, args, &result, &runtime.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: runtime.DefaultBackOffTimeAcrossRetries,
		},
	)
	if !assert.Error(t, err, "got tchannel error") {
		return
	}
	assert.False(t, success)
}
