package main

import (
	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func downloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "download",
		Run: download,
	}

	registerDatabaseFlags(cmd.Flags())
	registerStorageFlags(cmd.Flags())

	return cmd
}

func download(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := zap.L()

	group, ctx := errgroup.WithContext(ctx)
	storage := setupStorage(logger.Named("storage"))

	db, err := setupDatabase(ctx, logger)
	if err != nil {
		logger.Fatal("Unable to initialize the database", zap.Error(err))
	}

	var streams []radiochatter.Stream
	if err := db.Find(&streams).Error; err != nil {
		logger.Fatal("Unable to load the streams", zap.Error(err))
	}

	for _, stream := range streams {
		radiochatter.StartProcessing(ctx, logger, group, stream, storage, db)
	}

	defer logger.Info("Exit")

	if err := group.Wait(); err != nil {
		logger.Fatal("Failed", zap.Error(err))
	}
}
