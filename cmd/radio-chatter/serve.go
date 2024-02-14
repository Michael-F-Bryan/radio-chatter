package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/handlers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Addr string

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the GraphQL server",
		Run:   serve,
	}

	registerStorageFlags(cmd.Flags())
	registerDatabaseFlags(cmd.Flags())

	cmd.Flags().StringP("host", "H", "127.0.0.1", "The host to serve on")
	_ = viper.BindPFlag("serve.host", cmd.Flags().Lookup("host"))
	_ = viper.BindEnv("serve.host", "HOST")

	cmd.Flags().Uint16P("port", "p", 8080, "The port to serve on")
	_ = viper.BindPFlag("serve.port", cmd.Flags().Lookup("port"))
	_ = viper.BindEnv("serve.port", "PORT")

	return cmd
}

func serve(cmd *cobra.Command, args []string) {
	logger := zap.L()
	ctx := cmd.Context()
	cfg := LoadConfig()

	storage := setupStorage(logger, cfg.Storage)
	db := setupDatabase(ctx, logger, cfg)
	addr := cfg.Serve.Addr()

	server := http.Server{
		Addr:    addr,
		Handler: handlers.Router(logger, db, storage),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Fatal("Graceful shutdown failed", zap.Error(err))
		}
	}()

	logger.Info("Starting server", zap.String("addr", addr))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal("Serving failed", zap.Error(err))
	}
}
