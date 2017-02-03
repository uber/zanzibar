package generated

import (
	"io/ioutil"
	"net/http"

	"code.uber.internal/example/example-gateway"
	generatedClient "code.uber.internal/example/example-gateway/clients/generated"

	"github.com/pkg/errors"
)

func HandleFooRequest(
	inc *gateway.IncomingMessage,
	gateway *gateway.EdgeGateway,
) {
	rawBody, ok := inc.ReadAll()
	if !ok {
		return
	}

	var body generatedClient.Foo
	if ok := inc.UnmarshalBody(&body, rawBody); !ok {
		return
	}

	h := make(http.Header)
	h.Set("x-uber-uuid", inc.Header.Get("x-uber-uuid"))

	clientBody := convertToClient(&body)
	clientResp, err := g.Clients.generated.Generated(&body, h)
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
	inc.CopyJSON(res.Res.StatusCode, res.Res.Body)
}

func convertToClient(
	body *Foo,
) *generatedClient.Generated {
	clientBody := &generatedClient.Generated
    // TODO(sindelar): Add field mappings here.
    return body
	}
}
