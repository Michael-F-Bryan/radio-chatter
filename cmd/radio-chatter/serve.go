package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/handlers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var Addr string

func serveCmd() *cobra.Command {
	var port uint16
	var host string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the GraphQL server",
		Run:   func(cmd *cobra.Command, args []string) { serve(cmd.Context(), fmt.Sprintf("%s:%d", host, port)) },
	}

	registerDatabaseFlags(cmd.Flags())

	cmd.Flags().StringVarP(&host, "host", "H", "127.0.0.1", "The host to serve on")
	_ = viper.BindPFlag("serve.host", cmd.Flags().Lookup("host"))
	cmd.Flags().Uint16VarP(&port, "port", "p", 8080, "The port to serve on")
	_ = viper.BindPFlag("serve.port", cmd.Flags().Lookup("port"))

	return cmd
}

func serve(ctx context.Context, addr string) {
	logger := zap.L()

	db := setupDatabase(ctx, logger)

	server := http.Server{
		Addr:    addr,
		Handler: handlers.Router(logger, db),
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
