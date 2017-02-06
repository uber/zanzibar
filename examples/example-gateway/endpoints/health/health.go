package health

import (
	"github.com/uber/zanzibar/examples/example-gateway/clients"
	"github.com/uber/zanzibar/runtime"
)

// JSONResponse ...
type JSONResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}

// HandleHealthRequest for the health request
func HandleHealthRequest(
	inc *zanzibar.IncomingMessage,
	g *zanzibar.Gateway,
	clients *clients.Clients,
) {
	resp := &JSONResponse{
		Ok:      true,
		Message: "Healthy, from example-gateway",
	}

	inc.WriteJSON(200, resp)
}
