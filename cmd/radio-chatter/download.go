package main

import (
	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func downloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download and archive all streams",
		Run:   download,
	}

	registerDatabaseFlags(cmd.Flags())
	registerStorageFlags(cmd.Flags())

	return cmd
}

func download(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := zap.L()
	cfg := GetConfig(ctx)

	group, ctx := errgroup.WithContext(ctx)
	storage := setupStorage(logger, cfg.Storage)
	defer storage.Close()
	db := setupDatabase(ctx, logger, cfg)

	var streams []radiochatter.Stream
	if err := db.Find(&streams).Error; err != nil {
		logger.Fatal("Unable to load the streams", zap.Error(err))
	}

	for _, stream := range streams {
		cleanup := radiochatter.StartProcessing(ctx, logger, group, stream, storage, db)
		defer cleanup()
	}

	defer logger.Info("Exit")

	if err := group.Wait(); err != nil {
		logger.Fatal("Failed", zap.Error(err))
	}
}
