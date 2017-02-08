package googleNow

import (
	"context"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/uber/zanzibar/examples/example-gateway/clients"
	googleNowClient "github.com/uber/zanzibar/examples/example-gateway/clients/google_now"
	"github.com/uber/zanzibar/runtime"
)

// HandleAddCredentials handles /googlenow/add-credentials
func HandleAddCredentials(
	ctx context.Context,
	inc *zanzibar.IncomingMessage,
	g *zanzibar.Gateway,
	clients *clients.Clients,
) {
	rawBody, ok := inc.ReadAll()
	if !ok {
		return
	}

	// TODO(zw): use the request type generated from zanzibar endpoint config.
	var body googleNowClient.AddCredentialRequest
	if ok := inc.UnmarshalBody(&body, rawBody); !ok {
		return
	}

	h := make(http.Header)
	h.Set("x-uuid", inc.Header.Get("x-uuid"))

	clientResp, err := clients.GoogleNow.AddCredential(ctx, &body, h)
	if err != nil {
		inc.SendError(500, errors.Wrap(err, "could not make client request:"))
		return
	}

	defer func() {
		if err := clientResp.Body.Close(); err != nil {
			inc.SendError(500, errors.Wrap(err, "could not close client response body:"))
			return
		}
	}()
	b, err := ioutil.ReadAll(clientResp.Body)
	if err != nil {
		inc.SendError(500, errors.Wrap(err, "could not read client response body:"))
		return
	}

	if !isOKResponse(clientResp.StatusCode, []int{200, 202, 204}) {
		inc.SendErrorString(clientResp.StatusCode, string(b))
		return
	}

	// No need to unmarshal here because we expect empty body being returned.
	// Add following code if the body is not empty.
	//
	// var clientRespBody rtnowClient.AddCredentialResponse
	// if err := clientRespBody.UnmarshalJSON(b); err != nil {
	// 	inc.SendError(500, errors.Wrap(err, "could not unmarshal client response body:"))
	// 	return
	// }

	// TODO(zw): map clientRespBody to endpoint response body if needed.
	// Here we take shortcut since the response body doesn't change.
	inc.WriteJSONBytes(clientResp.StatusCode, nil)
}

func isOKResponse(statusCode int, okResponses []int) bool {
	for _, r := range okResponses {
		if statusCode == r {
			return true
		}
	}
	return false
}
