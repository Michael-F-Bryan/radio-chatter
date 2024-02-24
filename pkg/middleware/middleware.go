// package middleware contains common middleware functions used by our HTTP
// servers.
package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
)

// applyMiddleware will wrap a handler in a series of middleware functions,
// taking care to make sure the middleware is executed in the order they are
// provided.
func Apply(handler http.Handler, middleware ...mux.MiddlewareFunc) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		m := middleware[i]
		handler = m(handler)
	}
	return handler
}
