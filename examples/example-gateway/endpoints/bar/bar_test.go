package bar_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/bar/bar"
	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients-idl/clients/bar/bar"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
	runtime "github.com/uber/zanzibar/runtime"
	benchGateway "github.com/uber/zanzibar/test/lib/bench_gateway"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
	"github.com/uber/zanzibar/test/lib/util"
)

var defaultTestOptions = &testGateway.Options{
	KnownHTTPBackends: []string{"bar"},
	ConfigFiles:       util.DefaultConfigFiles("example-gateway"),
}
var defaultTestConfig = map[string]interface{}{
	"clients.bar.serviceName": "bar",
}

func TestBarListAndEnumEndpoint(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ms := ms.MustCreateTestService(t)

	ms.Start()
	defer ms.Stop()

	demoType := bar.DemoTypeFirst
	var args interface{} = &bar.Bar_ListAndEnum_Args{
		DemoIds:  []string{"abc", "def"},
		DemoType: &demoType,
		Demos:    []bar.DemoType{bar.DemoTypeFirst, bar.DemoTypeSecond},
	}
	ctx := context.Background()
	ms.MockClients().Bar.EXPECT().ListAndEnum(gomock.Any(), gomock.Any(), args, gomock.Any()).Return(ctx, "demo", nil, nil).AnyTimes()

	res, err := ms.MakeHTTPRequest(
		"GET", "/bar/list-and-enum?demoIds=abc&demoIds=def&demoType=FIRST&demos=FIRST&demos=SECOND", nil, nil,
	)

	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, 200, res.StatusCode)
}

func TestBarListAndEnumClient(t *testing.T) {
	gateway, err := benchGateway.CreateGateway(
		defaultTestConfig,
		defaultTestOptions,
		exampleGateway.CreateGateway,
	)
	if !assert.NoError(t, err) {
		return
	}
	defer gateway.Close()

	bgateway := gateway.(*benchGateway.BenchGateway)

	bgateway.HTTPBackends()["bar"].HandleFunc(
		"GET", "/bar/list-and-enum",
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, []string{"abc", "def"}, r.URL.Query()["demoIds"])
			assert.Equal(t, "FIRST", r.URL.Query().Get("demoType"))
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`"test-response"`))
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	demoType := bar.DemoTypeFirst
	demos := []bar.DemoType{bar.DemoTypeFirst, bar.DemoTypeSecond}
	_, body, _, err := bClient.ListAndEnum(
		context.Background(),
		nil,
		&clientsBarBar.Bar_ListAndEnum_Args{
			DemoIds:  []string{"abc", "def"},
			DemoType: &demoType,
			Demos:    demos,
		},
		&runtime.TimeoutAndRetryOptions{
			OverallTimeoutInMs:           time.Duration(5000) * time.Millisecond,
			RequestTimeoutPerAttemptInMs: time.Duration(2000) * time.Millisecond,
			MaxAttempts:                  2,
			BackOffTimeAcrossRetriesInMs: runtime.DefaultBackOffTimeAcrossRetries,
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, body)

	logs := bgateway.AllLogs()
	assert.Len(t, logs["Finished an outgoing client HTTP request"], 1)
}
