package serverless

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	endpointServerless "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/serverless/serverless"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
)

func TestServerlessEndpointCall(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	var a = "Test"
	endpointRequest := &endpointServerless.Serverless_Beta_Args{
		Request: &endpointServerless.Request{
			FirstName: &a,
		},
	}
	rawBody, _ := endpointRequest.MarshalJSON()
	res, err := ms.MakeHTTPRequest(
		"POST", "/serverless/post-request", nil, bytes.NewReader(rawBody),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	respBytes, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, string("{\"firstName\":\"Test\"}"), string(respBytes))
	fmt.Println(res.Body)
}
