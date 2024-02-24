package middleware

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var ErrCantHijack = errors.New("websocket: response does not implement http.Hijacker")

type loggerKey struct{}

// GetLogger will extract the logger associated with the current context.
func GetLogger(ctx context.Context) *zap.Logger {
	logger, exists := ctx.Value(loggerKey{}).(*zap.Logger)
	if exists {
		return logger
	}

	return zap.L()
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// Logging is a middleware function that will attach a logger to the request
// context and log the result of executing that request.
func Logging(logger *zap.Logger) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()

			requestID, ok := GetRequestID(r.Context())
			if !ok {
				logger.DPanic("Logging middleware must be executed after the request ID middleware")
			}
			subLogger := logger.With(zap.String("request-id", requestID))
			ctx := WithLogger(r.Context(), subLogger)
			r = r.WithContext(ctx)

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
			subLogger.Debug("Response Headers", zap.Any("headers", writer.Header()))
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

func (s *spyResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := s.inner.(http.Hijacker); ok {
		return hijacker.Hijack()
	}

	return nil, nil, ErrCantHijack
}
