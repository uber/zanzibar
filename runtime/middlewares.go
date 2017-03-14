package middleware

type HandlerFn func(
	ctx context.Context,
	inc *zanzibar.IncomingMessage,
	gateway *zanzibar.Gateway,
	clients *clients.Clients,
)


type Adapter func(http.Handler) http.Handler

type Middleware interface {
    
}
