package middlewareFoo

import "net/http"

func middlewareFoo(next middlewares.HandlerFn) middlewares.HandlerFn {
	ctx.Put("token", "c9e452805dee5044ba520198628abcaa")
	next.ServeHTTP(w, r)
}

func buildMiddleware(string arg1) {
	return middlewareFoo
}

type Adapter func(http.Handler) http.Handler

func WithHeader(key, value string) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header.Add(key, value)
			h.ServeHTTP(w, r)
		})
	}
}
