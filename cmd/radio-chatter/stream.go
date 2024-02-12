package main

import (
	"context"
	"os"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func streamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Common stream operations",
	}

	registerDatabaseFlags(cmd.PersistentFlags())

	cmd.AddCommand(streamListCmd(), streamAddCmd(), streamRemoveCmd())

	return cmd
}

func streamListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all known streams",
		Run:   func(cmd *cobra.Command, args []string) { streamList(cmd.Context()) },
	}

	return cmd
}

func streamList(ctx context.Context) {
	logger := zap.L()

	db, err := setupDatabase(ctx, logger)
	if err != nil {
		logger.Fatal("Unable to initialize the database", zap.Error(err))
	}

	var streams []radiochatter.Stream

	if err := db.Find(&streams).Error; err != nil {
		logger.Fatal("Unable to load streams", zap.Error(err))
	}

	if err := Format.Print(os.Stdout, streams); err != nil {
		logger.Fatal("Unable to print the streams", zap.Error(err))
	}
}

func streamAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new stream to the database",
		Run:   streamAdd,
		Args:  cobra.ExactArgs(2),
	}

	return cmd
}

func streamAdd(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := zap.L()

	db, err := setupDatabase(ctx, logger)
	if err != nil {
		logger.Fatal("Unable to initialize the database", zap.Error(err))
	}

	displayName := args[0]
	url := args[1]

	stream := radiochatter.Stream{
		DisplayName: displayName,
		Url:         url,
	}

	if err := db.Save(&stream).Error; err != nil {
		logger.Fatal(
			"Unable to save the stream",
			zap.Any("stream", stream),
			zap.Error(err),
		)
	}

	if err := Format.Print(os.Stdout, &stream); err != nil {
		logger.Fatal(
			"Unable to print the stream",
			zap.Any("stream", stream),
			zap.Error(err),
		)
	}
}

func streamRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete a stream from the database",
		Run:   streamRemove,
		Args:  cobra.ExactArgs(1),
	}

	return cmd
}

func streamRemove(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := zap.L()

	db, err := setupDatabase(ctx, logger)
	if err != nil {
		logger.Fatal("Unable to initialize the database", zap.Error(err))
	}

	displayName := args[0]

	stream := radiochatter.Stream{DisplayName: displayName}

	if err := db.Delete(&stream).Error; err != nil {
		logger.Fatal(
			"Unable to delete the stream",
			zap.Error(err),
		)
	}

	_ = Format.Print(os.Stdout, stream)
}
