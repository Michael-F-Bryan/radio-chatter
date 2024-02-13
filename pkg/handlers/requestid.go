package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type requestIDKey struct{}

// GetRequestID will extract the request ID from the current context.
func GetRequestID(ctx context.Context) (id string, exists bool) {
	id, exists = ctx.Value(requestIDKey{}).(string)
	return id, exists
}

func requestIDMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		r = r.WithContext(ctx)

		headers := w.Header()
		if headers.Get("X-Request-Id") == "" {
			w.Header().Set("X-Request-Id", requestID)
		}

		h.ServeHTTP(w, r)
	})
}
