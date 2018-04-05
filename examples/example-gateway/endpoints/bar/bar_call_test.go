package bar_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	clientsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/bar/bar"
	endpointsBarBar "github.com/uber/zanzibar/examples/example-gateway/build/gen-code/endpoints/bar/bar"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestBarCall(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	userUUID := "1-1-1-1"
	expectedRequest := &clientsBarBar.Bar_ArgWithQueryParams_Args{
		Name:     "kot",
		UserUUID: &userUUID,
		// This is currently broken due to https://github.com/uber/zanzibar/issues/371
		// Tags:     []string{"foo", "bar"},
		Tags: []string{},
	}
	expectedHeaders := map[string]string{}
	barClientResp := &clientsBarBar.BarResponse{
		StringField: "foobar",
		BinaryField: []byte(""),
	}

	ms.MockClients().Bar.EXPECT().ArgWithQueryParams(gomock.Any(), expectedHeaders, expectedRequest).
		Return(barClientResp, map[string]string{}, nil)

	resp, err := ms.MakeHTTPRequest("GET", "/bar/argWithQueryParams?tags=foo&tags=bar&userUUID=1-1-1-1&name=kot", map[string]string{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	respBody := &endpointsBarBar.BarResponse{}
	err = json.Unmarshal(respBytes, respBody)
	assert.NoError(t, err)

	// t.Logf("Resp body: %s", respBytes)
	// t.Logf("Resp body: %+v", respBody)

	assert.Equal(t, "foobar", respBody.StringField)
}
