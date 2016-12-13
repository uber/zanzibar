package endpoints

import (
	"github.com/uber/zanzibar/examples/example-gateway/clients"
	"github.com/uber/zanzibar/examples/example-gateway/endpoints/contacts"
	"github.com/uber/zanzibar/examples/example-gateway/endpoints/google_now"
	"github.com/uber/zanzibar/examples/example-gateway/endpoints/health"
	"github.com/uber/zanzibar/runtime"
)

type handlerFn func(
	inc *zanzibar.IncomingMessage,
	gateway *zanzibar.Gateway,
	clients *clients.Clients,
)

type myEndpoint struct {
	HandlerFn handlerFn
	Clients   *clients.Clients
}

func (endpoint *myEndpoint) handle(
	inc *zanzibar.IncomingMessage, g *zanzibar.Gateway,
) {
	fn := endpoint.HandlerFn
	fn(inc, g, endpoint.Clients)
}

func makeEndpoint(
	g *zanzibar.Gateway,
	endpointName string,
	handlerName string,
	handlerFn handlerFn,
) *zanzibar.Endpoint {
	myEndpoint := &myEndpoint{
		Clients:   g.Clients.(*clients.Clients),
		HandlerFn: handlerFn,
	}

	return zanzibar.NewEndpoint(
		g,
		endpointName,
		handlerName,
		myEndpoint.handle,
	)
}

// Register will register all endpoints
func Register(g *zanzibar.Gateway, router *zanzibar.Router) {
	router.Register(
		"POST", "/contacts/:userUUID/contacts",
		makeEndpoint(
			g,
			"contacts",
			"saveContacts",
			contacts.HandleSaveContactsRequest,
		),
	)

	router.Register(
		"GET", "/health",
		makeEndpoint(
			g,
			"health",
			"health",
			health.HandleHealthRequest,
		),
	)

	router.Register(
		"POST", "/googlenow/add-credentials",
		makeEndpoint(
			g,
			"googlenow",
			"addCredentials",
			googleNow.HandleAddCredentials,
		),
	)
}
