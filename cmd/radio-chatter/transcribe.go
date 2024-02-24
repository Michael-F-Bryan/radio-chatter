package main

import (
	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func transcribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transcribe",
		Short: "Run speech-to-text on any radio messages that have been detected",
		Run:   transcribe,
	}
	registerDatabaseFlags(cmd.Flags())
	registerStorageFlags(cmd.Flags())
	return cmd
}

func transcribe(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := zap.L()
	cfg := GetConfig(ctx)
	db := setupDatabase(ctx, logger, cfg)
	storage := setupStorage(logger, cfg.Storage)
	defer storage.Close()
	whisper := radiochatter.NewWhisperTranscriber(logger.Named("whisper"))

	logger.Info("Started running speech-to-text")

	if err := radiochatter.Transcribe(ctx, logger.Named("transcribe"), db, whisper, storage); err != nil {
		logger.Fatal("Transcription failed", zap.Error(err))
	}
}
