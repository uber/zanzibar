package panichandler_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints-idl/endpoints/tchannel/baz/baz"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
	runtime "github.com/uber/zanzibar/runtime"
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
		ctx, "SimpleService", "AnotherCall", reqHeaders, args, &result, &runtime.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(3000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  1,
			BackOffTimeAcrossRetriesInMs: runtime.DefaultBackOffTimeAcrossRetries,
		},
	)

	assert.Error(t, err, "got tchannel error")
}
