package radiochatter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MaxSpeechToTextBatchSize is the maximum number of messages we'll try to
// transcribe at a time.
const MaxSpeechToTextBatchSize = 50
const DefaultWhisperModel = "large-v2"
const whisperCommand = "whisper"

type SpeechToText interface {
	// Transcribe the audio files located at the provided URLs into english.
	SpeechToText(ctx context.Context, urls []*url.URL) ([]string, error)
	// How many audio files can be translated in a single batch.
	//
	// You can pass more than this number to SpeechToText(), but the
	// implementation may break the batch into chunks internally.
	MaxBatchSize() int
}

type transcriber struct {
	logger  *zap.Logger
	db      *gorm.DB
	stt     SpeechToText
	storage blob.Storage
}

// Transcribe will continuously poll the database for new messages and run
// speech-to-text on them.
func Transcribe(ctx context.Context, logger *zap.Logger, db *gorm.DB, stt SpeechToText, storage blob.Storage) error {
	t := transcriber{
		logger:  logger,
		db:      db.WithContext(ctx),
		stt:     stt,
		storage: storage,
	}

	// Note: there's no point polling more rapidly than chunks are generated
	ticker := time.NewTicker(ChunkLength)
	defer ticker.Stop()

	for {
		// Clear out the backlog. We do this syncronously because it lets us
		// provide backpressure if the upstream service is too slow.
		for {
			numTranscribed, err := t.transcribeOnce(ctx)
			if errors.Is(err, context.Canceled) {
				// We don't count cancellation as an error
				return nil
			} else if err != nil {
				return err
			}

			if numTranscribed == 0 {
				// Looks like we're caught up
				break
			}
		}

		// Wait until we need to do our next run
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(ChunkLength):
		}
	}
}

func (t *transcriber) transcribeOnce(ctx context.Context) (int, error) {
	transmissions, err := untranscribedTransmissions(t.db, t.stt.MaxBatchSize())
	if err != nil {
		return 0, err
	} else if len(transmissions) == 0 {
		// Nothing to do...
		return 0, nil
	}

	t.logger.Debug("Transcribing", zap.Any("transmissions", transmissions))

	var urls []*url.URL

	for _, transmission := range transmissions {
		key, err := blob.ParseKey(transmission.Sha256)
		if err != nil {
			return 0, fmt.Errorf("unable to parse %q as a blob key: %w", transmission.Sha256, err)
		}
		url, err := t.storage.Link(ctx, key, 1*time.Hour)
		if err != nil {
			return 0, fmt.Errorf("unable to get a link to %q: %w", key, err)
		}
		urls = append(urls, url)
	}

	transcriptions, err := t.stt.SpeechToText(ctx, urls)
	if err != nil {
		return 0, fmt.Errorf("transcription failed: %w", err)
	} else if len(transcriptions) != len(urls) {
		return 0, fmt.Errorf("transcriber returned %d strings, but expected %d", len(transcriptions), len(urls))
	}

	var models []Transcription

	for i := 0; i < len(transcriptions); i++ {
		transmission := transmissions[i]
		model := Transcription{
			Content:        transcriptions[i],
			TransmissionID: transmission.ID,
		}
		models = append(models, model)
	}

	if err := t.db.Save(&models).Error; err != nil {
		return 0, fmt.Errorf("unable to save the new transcriptions: %w", err)
	}

	t.logger.Info("Saved transcriptions", zap.Any("transcriptions", models))

	return len(models), nil
}

// untranscribedTransmissions will query the database for the next batch of
// Transmissions that need to be transcribed.
func untranscribedTransmissions(db *gorm.DB, maxBatchSize int) ([]Transmission, error) {
	var transmissions []Transmission
	err := db.Joins("LEFT JOIN transcriptions ON transcriptions.transmission_id = transmissions.id").
		Where("transcriptions.id IS NULL").
		Limit(maxBatchSize).
		Find(&transmissions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query untranscribed transmissions: %w", err)
	}

	return transmissions, nil
}

type WhisperTranscriber struct {
	logger *zap.Logger
	model  string
}

func NewWhisperTranscriber(logger *zap.Logger) *WhisperTranscriber {
	return &WhisperTranscriber{
		logger: logger,
		model:  DefaultWhisperModel,
	}
}

func (w WhisperTranscriber) MaxBatchSize() int {
	return 1
}

func (w WhisperTranscriber) SpeechToText(ctx context.Context, urls []*url.URL) ([]string, error) {
	var results []string

	for _, url := range urls {
		text, err := w.transcribe(ctx, url)
		if err != nil {
			return nil, err
		}
		results = append(results, text)
	}

	return results, nil
}

func (w WhisperTranscriber) transcribe(ctx context.Context, url *url.URL) (string, error) {
	logger := w.logger.With(zap.Stringer("url", url))
	start := time.Now()

	f, cleanup, err := downloadUrl(ctx, logger, url)
	if err != nil {
		return "", fmt.Errorf("unable to download %s: %w", url, err)
	}
	defer cleanup()
	defer f.Close()

	tmp, err := os.MkdirTemp("", "radio-chatter-whisper-tmp*")
	if err != nil {
		return "", fmt.Errorf("unable to create a temp directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			logger.Warn(
				"Unable to clean up the temporary directory",
				zap.String("tmp", tmp),
				zap.Error(err),
			)
		}
	}()

	args := []string{
		"--model", w.model, "--language", "en", "--output_dir", tmp,
		"--output_format", "all", f.Name(),
	}

	cmd := exec.CommandContext(ctx, whisperCommand, args...)

	cmd.WaitDelay = 10 * time.Second
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }

	stdout := bytes.Buffer{}
	cmd.Stdout = &stdout
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr

	logger.Debug("Running whisper", zap.Stringer("cmd", cmd))
	whisperStarted := time.Now()

	err = cmd.Run()

	if commandWasCancelled(ctx, err) {
		return "", context.Canceled
	} else if err != nil {
		logger.Warn(
			"Whisper failed",
			zap.Stringer("stdout", &stdout),
			zap.Stringer("stderr", &stderr),
			zap.Int("exit-code", cmd.ProcessState.ExitCode()),
		)
		return "", fmt.Errorf("transcription with Whisper failed: %w", err)
	}

	// Note: If we were transcribing path/to/whatever.mp3, the transcription
	// would be saved as $tmp/whatever.txt
	filename := filepath.Base(f.Name())
	filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".txt"
	fullPath := path.Join(tmp, filename)

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("unable to read %q: %w", fullPath, err)
	}

	end := time.Now()

	logger.Debug(
		"Finished transcribing",
		zap.ByteString("transcription", content),
		zap.Duration("total-duration", end.Sub(start)),
		zap.Duration("whisper-duration", end.Sub(whisperStarted)),
	)

	return string(content), nil
}

func downloadUrl(ctx context.Context, logger *zap.Logger, url *url.URL) (*os.File, func(), error) {
	dummyCancel := func() {}

	if url.Scheme == "file" {
		f, err := os.Open(url.Path)
		// Note: we don't need to do any cleanup because it's an existing file
		return f, dummyCancel, err
	}

	f, err := os.CreateTemp("", "radio-chatter-tmp*")
	if err != nil {
		return nil, dummyCancel, err
	}

	logger.Debug("Downloading")
	response, err := http.Get(url.String())
	if err != nil {
		return nil, dummyCancel, err
	}
	defer response.Body.Close()
	logger.Debug(
		"Reading response",
		zap.Int("status", response.StatusCode),
		zap.Int64("content-length", response.ContentLength),
		zap.Any("headers", response.Header),
	)

	if response.StatusCode < 200 || response.StatusCode >= 400 {
		return nil, dummyCancel, errors.New(response.Status)
	}

	bytesWritten, err := io.Copy(f, response.Body)
	if err != nil {
		return nil, dummyCancel, fmt.Errorf("unable to read the response: %s", response.Status)
	}

	if err = f.Sync(); err != nil {
		return nil, dummyCancel, fmt.Errorf("flushing %q failed: %w", f.Name(), err)
	}

	logger.Debug(
		"Response downloaded",
		zap.Int64("bytes-written", bytesWritten),
		zap.String("tmp", f.Name()),
	)

	cleanup := func() {
		path := f.Name()
		if err := os.Remove(path); err != nil {
			logger.Warn("Unable to clean up temporary file", zap.String("path", path), zap.Error(err))
		}
	}

	return f, cleanup, nil
}

func commandWasCancelled(ctx context.Context, err error) bool {
	var exitError *exec.ExitError

	if !errors.As(err, &exitError) {
		// The error wasn't caused by a command exiting unsuccessfully
		return false
	}

	if exitError.ProcessState.ExitCode() >= 0 {
		// It exited early
		return false
	}

	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
