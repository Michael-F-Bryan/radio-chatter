package radiochatter

import (
	"context"
	"os"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

type thunk = func() error

func StartProcessing(
	ctx context.Context,
	logger *zap.Logger,
	group *errgroup.Group,
	stream Stream,
	storage blob.Storage,
	db *gorm.DB,
) (cleanup func()) {
	archiveOps := make(chan ArchiveOperation)

	temp, cleanup := mkdtemp(logger)
	logger.Debug(
		"Saving clips to a temporary directory",
		zap.String("path", temp),
		zap.Uint("stream-id", stream.ID),
		zap.String("stream-name", stream.DisplayName),
	)

	group.Go(preprocess(ctx, logger.Named("preprocess"), stream.Url, temp, archiveOps))
	group.Go(archive(ctx, logger.Named("archive"), archiveOps, storage, db, stream))

	return cleanup
}

func archive(
	ctx context.Context,
	logger *zap.Logger,
	archiveOps <-chan ArchiveOperation,
	storage blob.Storage,
	db *gorm.DB,
	stream Stream,
) thunk {
	return func() error {
		state := ArchiveState{
			Logger:  logger,
			Storage: storage,
			DB:      db.WithContext(ctx),
			Stream:  stream,
		}

		for {
			select {
			case op, ok := <-archiveOps:
				if !ok {
					// Channel was closed. No more ops to execute.
					return nil
				}

				logger.Debug("executing", zap.Reflect("op", op))

				if err := op.Execute(ctx, state); err != nil {
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
	dir string,
	archiveOps chan<- ArchiveOperation,
) thunk {
	return func() error {
		defer close(archiveOps)

		cb := ArchiveCallbacks(ctx, archiveOps)
		return Preprocess(ctx, logger, url, dir, cb)
	}
}

func mkdtemp(logger *zap.Logger) (string, context.CancelFunc) {
	dir, err := os.MkdirTemp("", "radio-chatter-tmp")
	if err != nil {
		logger.Fatal("Unable to create a temporary directory", zap.Error(err))
	}
	cancel := func() {
		if err := os.RemoveAll(dir); err != nil {
			logger.Error(
				"Unable to remove the temporary directory",
				zap.String("temp", dir),
				zap.Error(err),
			)
		}
	}

	return dir, cancel
}
