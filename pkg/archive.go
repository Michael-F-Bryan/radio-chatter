package radiochatter

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// ArchiveState is the state passed to archive operations.
type ArchiveState struct {
	Logger  *zap.Logger
	Storage BlobStorage
}

// ArchiveCallbacks gets a set of PreprocessingCallbacks that will send archiver
// operations down a channel in response to preprocessing events.
func ArchiveCallbacks(ctx context.Context, ch chan<- ArchiveOperation) PreprocessingCallbacks {
	return archiveCallbacks(ctx, ch, time.Now)
}

func archiveCallbacks(ctx context.Context, ch chan<- ArchiveOperation, now func() time.Time) PreprocessingCallbacks {
	a := archiver{
		ctx: ctx,
		ch:  ch,
		now: now,
	}

	cb := PreprocessingCallbacks{
		DownloadStarted: a.onDownloadStarted,
		StartWriting:    a.onStartWriting,
		SilenceStart:    a.onSilenceStart,
		SilenceEnd:      a.onSilenceEnd,
		Finished:        a.onFinished,
	}

	return cb
}

type archiver struct {
	ch  chan<- ArchiveOperation
	ctx context.Context
	now func() time.Time

	currentFile      string
	recordingStarted time.Time
	fileIndex        int
	inSilence        bool
	audioStarted     time.Duration
	spans            []audioSpan
}

func (a *archiver) emit(op ArchiveOperation) {
	select {
	case a.ch <- op:
	case <-a.ctx.Done():
	}
}

func (a *archiver) onDownloadStarted() {
	a.recordingStarted = a.now()
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
		Start: a.audioStarted - startOffset,
		End:   t - startOffset,
	}

	// Note: We want to ignore tiny spans of audio
	if span.End-span.Start > 10*time.Millisecond {
		a.spans = append(a.spans, span)
	}

	a.inSilence = true
}

func (a *archiver) onSilenceEnd(t time.Duration, duration time.Duration) {
	a.audioStarted = t
	a.inSilence = false
}

func (a *archiver) completeFile(audioMayContinue bool) {
	startOffset := clipLength * time.Duration(a.fileIndex)
	clipStart := a.recordingStarted.Add(startOffset).UTC()

	a.emit(saveChunk{Path: a.currentFile, Timestamp: clipStart})

	if !a.inSilence && audioMayContinue {
		// Make sure we handle audio that continues across the end of the
		// current clip
		endOfChunk := startOffset + clipLength
		span := audioSpan{
			Start: a.audioStarted - startOffset,
			End:   endOfChunk - startOffset,
		}
		a.spans = append(a.spans, span)
		// Make sure the next audio clip doesn't include the bits we got
		a.audioStarted = endOfChunk
	}

	if a.spans != nil {
		op := splitAudioSnippets{
			Path:      a.currentFile,
			Timestamp: clipStart,
			Pieces:    a.spans,
		}
		a.emit(op)
		a.spans = nil
	}
}

type ArchiveOperation interface {
	fmt.Stringer
	Apply(ctx context.Context, state ArchiveState) error
}

// saveChunk takes a chunk of audio and saves it for later retrieval.
type saveChunk struct {
	Path string
	// When the chunk started.
	Timestamp time.Time
}

func (p saveChunk) String() string {
	return fmt.Sprintf("Save %q to blob storage", p.Path)
}

func (p saveChunk) Apply(ctx context.Context, state ArchiveState) error {
	data, err := os.ReadFile(p.Path)
	if err != nil {
		return fmt.Errorf("unable to read %q: %w", p.Path, err)
	}

	key, err := state.Storage.Store(ctx, data)
	if err != nil {
		return fmt.Errorf("unable to save %q to blob storage: %w", p.Path, err)
	}

	state.Logger.Info("Chunk saved to blob storage", zap.String("path", p.Path), zap.Stringer("key", key))

	return nil
}

type splitAudioSnippets struct {
	Path string
	// When the clip started.
	Timestamp time.Time
	Pieces    []audioSpan
}

func (s splitAudioSnippets) String() string {
	return fmt.Sprintf("Split %q into %d snippets of audio", s.Path, len(s.Pieces))
}

func (s splitAudioSnippets) Apply(ctx context.Context, state ArchiveState) error {
	state.Logger.Info(
		"Splitting",
		zap.String("path", s.Path),
		zap.Any("snippets", s.Pieces),
	)

	group, ctx := errgroup.WithContext(ctx)

	for _, piece := range s.Pieces {
		group.Go(splitAudio(ctx, state, s.Path, piece))
	}

	return group.Wait()
}

func splitAudio(ctx context.Context, state ArchiveState, path string, span audioSpan) func() error {
	return func() error {
		tmp := filepath.Join(os.TempDir(), fmt.Sprintf("split-%d.mp3", rand.Int63()))
		defer func() {
			if err := os.Remove(tmp); err != nil {
				state.Logger.Warn(
					"Unable to delete the temporary file",
					zap.String("path", tmp),
					zap.Error(err),
				)
			}
		}()

		buffer := 100 * time.Millisecond
		segmentStart := span.Start
		duration := span.Duration()

		if segmentStart > buffer {
			// Add a bit of space at the start and end of the clip so it doesn't
			// sound like it's been cut off.
			segmentStart -= buffer
			duration += 2 * buffer
		}

		segmentStart = segmentStart.Round(time.Millisecond)
		duration = duration.Round(time.Millisecond)

		args := []string{
			// Inputs
			"-i", path,
			// The segment start
			"-ss", fmt.Sprint(segmentStart.Seconds()),
			// Time duration
			"-t", fmt.Sprint(duration.Seconds()),
			// Reuse the same codec
			"-acodec", "copy",
			// Clean up the output so it's easier to troubleshoot
			"-hide_banner", "-nostdin", "-nostats",
			// We want to write output to our temporary file
			tmp,
		}

		cmd := exec.CommandContext(ctx, defaultCommand, args...)

		stdout := &bytes.Buffer{}
		cmd.Stdout = stdout
		stderr := &bytes.Buffer{}
		cmd.Stderr = stderr

		state.Logger.Debug("splitting with ffmpeg", zap.Stringer("cmd", cmd))

		err := cmd.Run()

		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				state.Logger.Warn(
					"ffmpeg errored out",
					zap.Stringer("cmd", cmd),
					zap.Int("code", exitError.ExitCode()),
					zap.ByteString("stderr", stderr.Bytes()),
					zap.ByteString("stdout", stdout.Bytes()),
				)
			}

			return fmt.Errorf("unable to extract %s from %q: %w", span, path, err)
		}

		buf, err := os.ReadFile(tmp)
		if err != nil {
			return fmt.Errorf("unable to read the split from %q: %w", tmp, err)
		}

		key, err := state.Storage.Store(ctx, buf)
		if err != nil {
			return fmt.Errorf("unable to store %s from %q: %w", span, path, err)
		}

		state.Logger.Info(
			"Wrote transmission to blob storage",
			zap.Stringer("key", key),
			zap.String("path", path),
			zap.Stringer("span", span),
			zap.Int("bytes", len(buf)),
		)

		return nil
	}
}

type audioSpan struct {
	Start time.Duration
	End   time.Duration
}

func (a audioSpan) String() string {
	return fmt.Sprintf("%s..%s", a.Start.Round(1*time.Millisecond), a.End.Round(1*time.Millisecond))
}

func (a audioSpan) Duration() time.Duration {
	return a.End - a.Start
}
