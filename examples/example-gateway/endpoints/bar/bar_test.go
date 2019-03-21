package contacts_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/uber/zanzibar/examples/example-gateway/build/gen-code/clients/bar/bar"
	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
)

func TestBarListAndEnum(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	ms := ms.MustCreateTestService(t)

	ms.Start()
	defer ms.Stop()

	demoType := bar.DemoTypeFirst
	var args interface{} = &bar.Bar_ListAndEnum_Args{DemoIds: []string{"abc", "def"}, DemoType: &demoType}
	ms.MockClients().Bar.EXPECT().ListAndEnum(gomock.Any(), gomock.Any(), args).Return("demo", nil, nil).AnyTimes()

	res, err := ms.MakeHTTPRequest(
		"GET", "/bar/list-and-enum?demoIds[]=abc&demoIds[]=def&demoType=0", nil, nil,
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, 200, res.StatusCode)
}
