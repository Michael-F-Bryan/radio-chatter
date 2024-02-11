package radiochatter

import (
	"context"
	"fmt"
	"os"
	"time"

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
	a := archiver{ch: ch}

	cb := PreprocessingCallbacks{
		StartWriting: a.onStartWriting,
		SilenceStart: a.onSilenceStart,
		SilenceEnd:   a.onSilenceEnd,
		Finished:     a.onFinished,
	}

	return cb
}

type archiver struct {
	currentFile  string
	fileIndex    int
	inSilence    bool
	audioStarted time.Duration
	spans        []audioSpan
	ch           chan<- ArchiveOperation
}

func (a *archiver) onStartWriting(path string) {
	if a.currentFile != "" {
		a.completeFile(true)
		a.fileIndex++
	}
	a.currentFile = path
}

func (a *archiver) onFinished() {
	if a.currentFile != "" {
		// Looks like we're finished... Make sure the last chunk gets
		// persisted, too
		a.completeFile(false)
	}
}

func (a *archiver) onSilenceStart(t time.Duration) {
	startOffset := clipLength * time.Duration(a.fileIndex)
	span := audioSpan{
		start: a.audioStarted - startOffset,
		end:   t - startOffset,
	}

	// Note: We want to ignore tiny spans of audio
	if span.end-span.start > 10*time.Millisecond {
		a.spans = append(a.spans, span)
	}

	a.inSilence = true
}

func (a *archiver) onSilenceEnd(t time.Duration, duration time.Duration) {
	a.audioStarted = t
	a.inSilence = false
}

func (a *archiver) completeFile(audioMayContinue bool) {
	a.ch <- saveChunk{a.currentFile}

	if !a.inSilence && audioMayContinue {
		// Make sure we handle audio that continues across the end of the
		// current clip
		startOffset := clipLength * time.Duration(a.fileIndex)
		endOfChunk := startOffset + clipLength
		println(a.audioStarted, startOffset, endOfChunk)
		span := audioSpan{
			start: a.audioStarted - startOffset,
			end:   endOfChunk - startOffset,
		}
		a.spans = append(a.spans, span)
		// Make sure the next audio clip doesn't include the bits we got
		a.audioStarted = endOfChunk
	}

	if a.spans != nil {
		a.ch <- splitAudioSnippets{
			path:   a.currentFile,
			pieces: a.spans,
		}
		a.spans = nil
	}
}

type ArchiveOperation interface {
	fmt.Stringer
	Apply(ctx context.Context, state ArchiveState) error
}

// saveChunk takes a chunk of audio and saves it for later retrieval.
type saveChunk struct {
	path string
}

func (p saveChunk) String() string {
	return fmt.Sprintf("Save %q to blob storage", p.path)
}

func (p saveChunk) Apply(ctx context.Context, state ArchiveState) error {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return fmt.Errorf("unable to read %q: %w", p.path, err)
	}

	key, err := state.Storage.Store(ctx, data)
	if err != nil {
		return fmt.Errorf("unable to save %q to blob storage: %w", p.path, err)
	}

	state.Logger.Info("Chunk saved to blob storage", zap.String("path", p.path), zap.Stringer("key", key))

	return nil
}

type splitAudioSnippets struct {
	path   string
	pieces []audioSpan
}

func (s splitAudioSnippets) String() string {
	return fmt.Sprintf("Split %q into %d snippets of audio", s.path, len(s.pieces))
}

func (s splitAudioSnippets) Apply(ctx context.Context, state ArchiveState) error {
	state.Logger.Info(
		"Splitting",
		zap.String("path", s.path),
		zap.Any("snippets", s.pieces),
	)
	return nil
}

type audioSpan struct {
	start time.Duration
	end   time.Duration
}
