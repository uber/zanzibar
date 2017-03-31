// Code generated by zanzibar
// @generated

package endpoints

import (
	"context"

	"github.com/uber/zanzibar/examples/example-gateway/build/clients"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/bar"
	"github.com/uber/zanzibar/examples/example-gateway/build/endpoints/googlenow"
	"github.com/uber/zanzibar/examples/example-gateway/endpoints/contacts"
	"github.com/uber/zanzibar/examples/example-gateway/middlewares/example"
	"github.com/uber/zanzibar/runtime/middlewares/logger"

	"github.com/uber/zanzibar/runtime"

	// TODO: (lu) remove this, to be generated
	"github.com/uber/zanzibar/examples/example-gateway/endpoints/baz"
)

type handlerFn func(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
	clients *clients.Clients,
)

type myEndpoint struct {
	HandlerFn handlerFn
	Clients   *clients.Clients
}

func (endpoint *myEndpoint) handle(
	ctx context.Context,
	req *zanzibar.ServerHTTPRequest,
	res *zanzibar.ServerHTTPResponse,
) {
	fn := endpoint.HandlerFn
	fn(ctx, req, res, endpoint.Clients)
}

func makeEndpoint(
	g *zanzibar.Gateway,
	endpointName string,
	handlerName string,
	middlewares []zanzibar.MiddlewareHandle,
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
		zanzibar.NewStack(
			middlewares,
			myEndpoint.handle,
		).Handle,
	)
}

// Register will register all endpoints
func Register(g *zanzibar.Gateway, router *zanzibar.Router) {
	router.Register(
		"POST", "/bar/arg-not-struct-path",
		makeEndpoint(
			g,
			"bar",
			"argNotStruct",
			nil,
			bar.HandleArgNotStructRequest,
		),
	)
	router.Register(
		"GET", "/bar/missing-arg-path",
		makeEndpoint(
			g,
			"bar",
			"missingArg",
			nil,
			bar.HandleMissingArgRequest,
		),
	)
	router.Register(
		"GET", "/bar/no-request-path",
		makeEndpoint(
			g,
			"bar",
			"noRequest",
			nil,
			bar.HandleNoRequestRequest,
		),
	)
	router.Register(
		"POST", "/bar/bar-path",
		makeEndpoint(
			g,
			"bar",
			"normal",
			[]zanzibar.MiddlewareHandle{
				example.NewMiddleWare(
					g,
					example.Options{
						Foo: "test",
					},
				),
				logger.NewMiddleWare(
					g,
					logger.Options{},
				),
			},
			bar.HandleNormalRequest,
		),
	)
	router.Register(
		"POST", "/bar/too-many-args-path",
		makeEndpoint(
			g,
			"bar",
			"tooManyArgs",
			nil,
			bar.HandleTooManyArgsRequest,
		),
	)
	router.Register(
		"POST", "/contacts/:userUUID/contacts",
		makeEndpoint(
			g,
			"contacts",
			"saveContacts",
			nil,
			contacts.HandleSaveContactsRequest,
		),
	)
	router.Register(
		"POST", "/googlenow/add-credentials",
		makeEndpoint(
			g,
			"googlenow",
			"addCredentials",
			nil,
			googlenow.HandleAddCredentialsRequest,
		),
	)
	router.Register(
		"POST", "/googlenow/check-credentials",
		makeEndpoint(
			g,
			"googlenow",
			"checkCredentials",
			nil,
			googlenow.HandleCheckCredentialsRequest,
		),
	)

	// TODO: (lu) remove below, to be generated
	router.Register(
		"POST", "/baz/call-path",
		makeEndpoint(
			g,
			"baz",
			"call",
			nil,
			baz.HandleCallRequest,
		),
	)
	router.Register(
		"GET", "/baz/simple-path",
		makeEndpoint(
			g,
			"baz",
			"simple",
			nil,
			baz.HandleSimpleRequest,
		),
	)
	router.Register(
		"GET", "/baz/simple-future-path",
		makeEndpoint(
			g,
			"baz",
			"simpleFuture",
			nil,
			baz.HandleSimpleFutureRequest,
		),
	)
}
