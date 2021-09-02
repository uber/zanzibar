package panic_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	ms "github.com/uber/zanzibar/examples/example-gateway/build/services/example-gateway/mock-service"
)

func TestServiceCFrontHello(t *testing.T) {
	ms := ms.MustCreateTestService(t)
	ms.Start()
	defer ms.Stop()

	res, err := ms.MakeHTTPRequest(
		"GET", "/multi/serviceC_f/hello", nil, bytes.NewReader([]byte("body")),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(res.Body)
	assert.True(t, strings.Contains(buf.String(), "Unexpected workflow panic, recovered at endpoint."))
	assert.Equal(t, "502 Bad Gateway", res.Status)
}
