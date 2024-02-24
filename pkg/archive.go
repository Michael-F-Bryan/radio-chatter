package radiochatter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// ArchiveState is the state passed to archive operations.
type ArchiveState struct {
	Logger  *zap.Logger
	Storage blob.Storage
	DB      *gorm.DB
	Stream  Stream
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
	startOffset := ChunkLength * time.Duration(a.fileIndex)
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
	startOffset := ChunkLength * time.Duration(a.fileIndex)
	clipStart := a.recordingStarted.Add(startOffset).UTC()

	op := ArchiveOperation{
		Path:      a.currentFile,
		Timestamp: clipStart,
	}

	if !a.inSilence && audioMayContinue {
		// Make sure we handle audio that continues across the end of the
		// current clip
		endOfChunk := startOffset + ChunkLength
		span := audioSpan{
			Start: a.audioStarted - startOffset,
			End:   endOfChunk - startOffset,
		}
		a.spans = append(a.spans, span)
		// Make sure the next audio clip doesn't include the bits we got
		a.audioStarted = endOfChunk
	}

	if a.spans != nil {
		op.Pieces = a.spans
		a.spans = nil
	}

	select {
	case a.ch <- op:
	case <-a.ctx.Done():
	}
}

type ArchiveOperation struct {
	Path string
	// When the chunk started.
	Timestamp time.Time
	Pieces    []audioSpan
}

func (a ArchiveOperation) Execute(ctx context.Context, state ArchiveState) error {
	data, err := os.ReadFile(a.Path)
	if err != nil {
		return fmt.Errorf("unable to read %q: %w", a.Path, err)
	}

	key, err := state.Storage.Store(ctx, data)
	if err != nil {
		return fmt.Errorf("unable to save %q to blob storage: %w", a.Path, err)
	}

	chunk := Chunk{
		TimeStamp: a.Timestamp,
		Sha256:    key.String(),
		StreamID:  state.Stream.ID,
	}
	if err := state.DB.Save(&chunk).Error; err != nil {
		return fmt.Errorf("unable to save the chunk for %q (%s): %w", a.Path, key, err)
	}

	state.Logger.Info(
		"Saved chunk",
		zap.String("path", a.Path),
		zap.Int("bytes", len(data)),
		zap.Any("chunk", chunk),
	)

	if chunk.Transmissions != nil {
		if err := splitChunk(ctx, state, a, chunk); err != nil {
			return err
		}
	}

	state.Logger.Debug("Deleting original chunk file", zap.String("path", a.Path))
	if err := os.Remove(a.Path); err != nil {
		return fmt.Errorf("unable to delete %q: %w", a.Path, err)
	}

	return nil
}

func splitChunk(ctx context.Context, state ArchiveState, a ArchiveOperation, chunk Chunk) error {
	state.Logger.Debug(
		"Splitting",
		zap.String("path", a.Path),
		zap.Any("snippets", a.Pieces),
	)

	group, ctx := errgroup.WithContext(ctx)

	for _, piece := range a.Pieces {
		group.Go(splitAudioJob(ctx, state, a.Path, piece, chunk))
	}

	if err := group.Wait(); err != nil {
		return fmt.Errorf("unable to split audio: %w", err)
	}

	return nil
}

func splitAudioJob(ctx context.Context, state ArchiveState, path string, span audioSpan, chunk Chunk) func() error {
	return func() error {
		_, err := splitAudio(ctx, state, path, span, chunk)
		return err
	}
}

func splitAudio(ctx context.Context, state ArchiveState, path string, span audioSpan, chunk Chunk) (Transmission, error) {
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

	cmd := exec.CommandContext(ctx, ffmpegCommand, args...)

	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr

	state.Logger.Debug("splitting with ffmpeg", zap.Stringer("cmd", cmd))

	err := cmd.Run()

	if commandWasCancelled(ctx, err) {
		return Transmission{}, context.Canceled
	} else if err != nil {
		var exitError *exec.ExitError

		if errors.As(err, &exitError) {
			state.Logger.Warn(
				"ffmpeg errored out",
				zap.Stringer("cmd", cmd),
				zap.Int("code", exitError.ExitCode()),
				zap.ByteString("stderr", stderr.Bytes()),
				zap.ByteString("stdout", stdout.Bytes()),
			)
		}

		return Transmission{}, fmt.Errorf("unable to extract %s from %q: %w", span, path, err)
	}

	buf, err := os.ReadFile(tmp)
	if err != nil {
		return Transmission{}, fmt.Errorf("unable to read the split from %q: %w", tmp, err)
	}

	key, err := state.Storage.Store(ctx, buf)
	if err != nil {
		return Transmission{}, fmt.Errorf("unable to store %s from %q: %w", span, path, err)
	}

	transmission := Transmission{
		TimeStamp: chunk.TimeStamp.Add(span.Start),
		Length:    span.Duration(),
		Sha256:    key.String(),
		ChunkID:   chunk.ID,
	}

	if err := state.DB.Save(&transmission).Error; err != nil {
		return Transmission{}, fmt.Errorf("unable to save transmission: %w", err)
	}

	state.Logger.Info(
		"Saved transmission",
		zap.Any("transmission", transmission),
		zap.Int("bytes", len(buf)),
	)

	return transmission, nil
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
