/*
 * CODE GENERATED AUTOMATICALLY
 * THIS FILE SHOULD NOT BE EDITED BY HAND
 */

package bar

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/uber-go/zap"
	"github.com/uber/zanzibar/examples/example-gateway/clients"
	zanzibar "github.com/uber/zanzibar/runtime"

	barClient "github.com/uber/zanzibar/examples/example-gateway/clients/bar"
	bar "github.com/uber/zanzibar/examples/example-gateway/gen-code/github.com/uber/zanzibar/endpoints/bar/bar"
)

// HandleBarRequest handles request
func HandleBarRequest(
	ctx context.Context,
	inc *zanzibar.IncomingMessage,
	gateway *zanzibar.Gateway,
	clients *clients.Clients,
) {
	rawBody, ok := inc.ReadAll()
	if !ok {
		return
	}

	var body bar.BarRequest
	if ok := inc.UnmarshalBody(&body, rawBody); !ok {
		return
	}

	h := make(http.Header)
	h.Set("x-uber-uuid", inc.Header.Get("x-uber-uuid"))

	clientBody := convertToClient(&body)
	clientResp, err := clients.Bar.Bar(clientBody, h)
	if err != nil {
		gateway.Logger.Error("Could not make client request",
			zap.String("error", err.Error()),
		)
		inc.SendError(500, errors.Wrap(err, "Could not make client request:"))
		return
	}

	defer func() {
		if err := clientResp.Body.Close(); err != nil {
			inc.SendError(500, errors.Wrap(err, "Could not close client response body:"))
			return
		}
	}()
	b, err := ioutil.ReadAll(clientResp.Body)
	if err != nil {
		inc.SendError(500, errors.Wrap(err, "Could not read client response body:"))
		return
	}

	if !isOKResponse(clientResp.StatusCode, []int{200, 202, 204}) {
		inc.SendErrorString(clientResp.StatusCode, string(b))
		return
	}

	// TODO(sindelar): Apply response filtering and translation.
	inc.CopyJSON(clientResp.StatusCode, clientResp.Body)
}

func convertToClient(
	body *bar.BarRequest,
) *barClient.BarHTTPRequest {
	// TODO(sindelar): Add field mappings here. Cannot rely
	// on Go 1.8 casting for all conversions.
	clientBody := &barClient.BarHTTPRequest{}
	return clientBody
}

func isOKResponse(statusCode int, okResponses []int) bool {
	for _, r := range okResponses {
		if statusCode == r {
			return true
		}
	}
	return false
}
