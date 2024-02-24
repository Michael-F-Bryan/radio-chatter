package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type HealthStatus struct {
	Ok       bool           `json:"ok"`
	Database DatabaseHealth `json:"db"`
}

type DatabaseHealth struct {
	Ok           bool    `json:"ok"`
	ResponseTime float64 `json:"response-time"`
	Error        string  `json:"error,omitempty"`
}

func Healthz(db *gorm.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := middleware.GetLogger(r.Context())
		dbHealth := databaseHealth(r.Context(), db)

		status := HealthStatus{
			Ok:       dbHealth.Ok,
			Database: dbHealth,
		}

		logger.Info("Health check", zap.Any("status", status))

		if status.Ok {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err := json.NewEncoder(w).Encode(&status); err != nil {
			logger.Error(
				"Unable to send back the status",
				zap.Any("status", status),
				zap.Error(err),
			)
		}
	})
}

func databaseHealth(ctx context.Context, db *gorm.DB) DatabaseHealth {
	started := time.Now()

	sqlDb, err := db.DB()
	if err != nil {
		return DatabaseHealth{
			Ok:           false,
			ResponseTime: 0,
			Error:        err.Error(),
		}
	}

	if err := sqlDb.PingContext(ctx); err != nil {
		return DatabaseHealth{
			Ok:           false,
			ResponseTime: time.Since(started).Seconds(),
			Error:        err.Error(),
		}
	}

	return DatabaseHealth{
		Ok:           true,
		ResponseTime: time.Since(started).Seconds(),
	}
}
