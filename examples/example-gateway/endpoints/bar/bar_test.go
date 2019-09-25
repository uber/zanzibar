package contacts_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/bar/bar"
	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/bar/bar"
	exampleGateway "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
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
	var args interface{} = &bar.Bar_ListAndEnum_Args{DemoIds: []string{"abc", "def"}, DemoType: &demoType}
	ms.MockClients().Bar.EXPECT().ListAndEnum(gomock.Any(), gomock.Any(), args).Return("demo", nil, nil).AnyTimes()

	res, err := ms.MakeHTTPRequest(
		"GET", "/bar/list-and-enum?demoIds=abc&demoIds=def&demoType=0", nil, nil,
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
			assert.Equal(t, "0", r.URL.Query().Get("demoType"))
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`"test-response"`))
		},
	)

	deps := bgateway.Dependencies.(*exampleGateway.DependenciesTree)
	bClient := deps.Client.Bar

	demoType := bar.DemoTypeFirst
	body, _, err := bClient.ListAndEnum(
		context.Background(), nil, &clientsBarBar.Bar_ListAndEnum_Args{DemoIds: []string{"abc", "def"}, DemoType: &demoType},
	)
	assert.NoError(t, err)
	assert.NotNil(t, body)
}
