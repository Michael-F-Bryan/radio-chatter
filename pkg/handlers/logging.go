package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type loggerKey struct{}

// GetLogger will extract the logger associated with the current context.
func GetLogger(ctx context.Context) *zap.Logger {
	logger, exists := ctx.Value(loggerKey{}).(*zap.Logger)
	if exists {
		return logger
	}

	return zap.L()
}

func withLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// loggingMiddleware is a middleware function that will attach a logger to the
// request context and log the result of executing that request.
func loggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()

			subLogger := logger

			if requestID, ok := GetRequestID(r.Context()); ok {
				subLogger = logger.With(zap.String("request-id", requestID))
				ctx := withLogger(r.Context(), subLogger)
				r = r.WithContext(ctx)
			} else {
				logger.DPanic("Logging middleware must be executed after the request ID middleware")
			}

			writer := spyResponseWriter{inner: w}

			subLogger.Debug("Request Headers", zap.Any("headers", r.Header))

			h.ServeHTTP(&writer, r)

			duration := time.Since(started)
			subLogger.Info(
				"Handled",
				zap.Int("code", writer.code),
				zap.Stringer("url", r.URL),
				zap.String("method", r.Method),
				zap.Int("bytes-written", writer.bytesWritten),
				zap.Duration("duration", duration),
				zap.String("remote-addr", r.RemoteAddr),
				zap.String("user-agent", r.UserAgent()),
			)
		})
	}
}

type spyResponseWriter struct {
	code         int
	bytesWritten int
	inner        http.ResponseWriter
}

func (s *spyResponseWriter) Header() http.Header {
	return s.inner.Header()
}

func (s *spyResponseWriter) WriteHeader(code int) {
	s.code = code
	s.inner.WriteHeader(code)
}

func (s *spyResponseWriter) Write(data []byte) (int, error) {
	bytesWritten, err := s.inner.Write(data)
	s.bytesWritten += bytesWritten
	return bytesWritten, err
}
