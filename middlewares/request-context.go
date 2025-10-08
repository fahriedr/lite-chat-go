package middlewares

import (
	"context"
	"net/http"
)

func RequestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "request", r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
