package middlewares

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func ZapRequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		fallback, _ := zap.NewDevelopment()
		logger = fallback
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			start := time.Now()
			next.ServeHTTP(lrw, r)

			route := mux.CurrentRoute(r)
			path := ""
			if route != nil {
				path, _ = route.GetPathTemplate()
			}

			if lrw.statusCode >= 400 {
				logger.Error("HTTP request",
					zap.String("method", r.Method),
					zap.String("path", path),
					zap.Int("status", lrw.statusCode),
					zap.Duration("duration", time.Since(start)),
				)
				return
			}

			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", path),
				zap.Int("status", lrw.statusCode),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
