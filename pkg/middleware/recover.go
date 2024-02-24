package middleware

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Recover adds a handler which will automatically catch panics and log them.
func Recover(logger *zap.Logger) mux.MiddlewareFunc {
	return handlers.RecoveryHandler(handlers.RecoveryLogger(&recoveryLogger{logger}))
}

type recoveryLogger struct {
	logger *zap.Logger
}

func (r *recoveryLogger) Println(values ...any) {
	var payload any

	if len(values) == 1 {
		payload = values[0]
	} else {
		payload = values
	}

	r.logger.Error("A panic occurred", zap.Any("payload", payload), zap.Stack("stack"))
}
