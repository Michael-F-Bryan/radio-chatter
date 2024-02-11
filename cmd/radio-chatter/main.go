package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

func main() {
	args := parseArgs()
	args.initializeLogger()
	logger := zap.L()
	defer func() { _ = logger.Sync() }()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)
	storage := setupStorage(logger.Named("storage"))
	archiveOps := make(chan radiochatter.ArchiveOperation)

	group.Go(preprocess(ctx, logger.Named("preprocess"), args.url, archiveOps))
	group.Go(archive(ctx, logger.Named("archive"), archiveOps, storage))

	defer logger.Info("Shutting down")

	if err := group.Wait(); err != nil {
		logger.Fatal("Failed", zap.Error(err))
	}
}

type Thunk = func() error

func archive(
	ctx context.Context,
	logger *zap.Logger,
	archiveOps <-chan radiochatter.ArchiveOperation,
	storage radiochatter.BlobStorage,
) Thunk {
	return func() error {
		state := radiochatter.ArchiveState{
			Logger:  logger,
			Storage: storage,
		}

		for {
			select {
			case op, ok := <-archiveOps:
				if !ok {
					// Channel was closed. No more ops to execute.
					return nil
				}

				if err := op.Apply(ctx, state); err != nil {
					return err
				}
			case <-ctx.Done():
				return nil
			}
		}
	}
}

func preprocess(
	ctx context.Context,
	logger *zap.Logger,
	url string,
	archiveOps chan<- radiochatter.ArchiveOperation,
) Thunk {
	return func() error {
		dir, err := os.MkdirTemp("", "radio-chatter-tmp")
		if err != nil {
			logger.Fatal("Unable to create a temporary directory", zap.Error(err))
		}
		logger.Debug("Saving clips to a temporary directory", zap.String("path", dir))
		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				logger.Error(
					"Unable to remove the temporary directory",
					zap.String("temp", dir),
					zap.Error(err),
				)
			}
		}()

		cb := radiochatter.ArchiveCallbacks(archiveOps)
		return radiochatter.Preprocess(ctx, logger, url, dir, cb)
	}
}

type args struct {
	devMode bool
	url     string
}

func parseArgs() args {
	var args args

	flag.BoolVar(&args.devMode, "dev", false, "Enable dev mode")
	flag.StringVar(&args.url, "url", radiochatter.FeedUrl, "The feed to download")
	flag.Parse()

	return args
}

func (a args) initializeLogger() {
	var cfg zap.Config

	if a.devMode {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Unable to initialize the logger: %e", err)
	}

	zap.ReplaceGlobals(logger)
}

func setupStorage(logger *zap.Logger) radiochatter.BlobStorage {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		logger.Error("Unable to get the user's cache directory", zap.Error(err))
	}

	baseDir := path.Join(cacheDir, "radio-chatter", "clips")

	return radiochatter.NewOnDiskStorage(logger, baseDir)
}
