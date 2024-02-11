package radiochatter

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// ArchiveState is the state passed to archive operations.
type ArchiveState struct {
	Logger  *zap.Logger
	Storage BlobStorage
}

// ArchiveCallbacks gets a set of PreprocessingCallbacks that will send archiver
// operations down a channel in response to preprocessing events.
func ArchiveCallbacks(ch chan<- ArchiveOperation) PreprocessingCallbacks {
	var currentFile string

	cb := PreprocessingCallbacks{
		StartWriting: func(path string) {
			if currentFile != "" {
				ch <- saveChunk{currentFile}
			}
			currentFile = path
		},
		Finished: func(err error) {
			if currentFile != "" {
				// Looks like we're finished... Make sure the last chunk gets
				// persisted, too
				ch <- saveChunk{currentFile}
				currentFile = ""
			}
		},
	}

	return cb
}

type ArchiveOperation interface {
	fmt.Stringer
	Apply(ctx context.Context, state ArchiveState) error
}

// saveChunk takes a 60-second chunk of audio and saves it for later retrieval.
type saveChunk struct {
	path string
}

func (p saveChunk) String() string {
	return fmt.Sprintf("Save %q to blob storage", p.path)
}

func (p saveChunk) Apply(ctx context.Context, state ArchiveState) error {
	state.Logger.Debug("Persisting file", zap.String("path", p.path))

	data, err := os.ReadFile(p.path)
	if err != nil {
		return fmt.Errorf("unable to read %q: %w", p.path, err)
	}

	key, err := state.Storage.Store(ctx, data)
	if err != nil {
		return fmt.Errorf("unable to save %q to blob storage: %w", p.path, err)
	}

	state.Logger.Info("Saved file", zap.String("path", p.path), zap.Stringer("key", key))

	return nil

}
