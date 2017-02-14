/*
 * CODE GENERATED AUTOMATICALLY
 * THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package bar_test

import (
	"bytes"
	"net/http"
	"testing"

	assert "github.com/stretchr/testify/assert"
	config "github.com/uber/zanzibar/examples/example-gateway/config"
	testGateway "github.com/uber/zanzibar/test/lib/test_gateway"
)

var benchBytes = []byte("{\"testrequest\"}")

func TestBar(t *testing.T) {
	var counter int = 0

	config := &config.Config{}
	gateway, err := testGateway.CreateGateway(t, config, nil)
	if !assert.NoError(t, err, "got bootstrap err") {
		return
	}
	defer gateway.Close()

	bar := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		if _, err := w.Write([]byte("{\"statusCode\":200}")); err != nil {
			t.Fatal("can't write fake response")
		}
		counter++
	}
	gateway.Backends()["GoogleNow"].HandleFunc(
		"POST", "/add-credentials", bar,
	)

	res, err := gateway.MakeRequest(
		"POST", "/googlenow/add-credentials", bytes.NewReader(benchBytes),
	)
	if !assert.NoError(t, err, "got http error") {
		return
	}

	assert.Equal(t, "200 OK", res.Status)
	assert.Equal(t, 1, counter)
}
