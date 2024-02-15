package radiochatter

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MaxSpeechToTextBatchSize is the maximum number of messages we'll try to
// transcribe at a time.
const MaxSpeechToTextBatchSize = 50

type SpeechToText interface {
	// Transcribe the audio files located at the provided URLs into english.
	SpeechToText(ctx context.Context, urls []*url.URL) ([]string, error)
}

type Transcriber struct {
	logger  *zap.Logger
	db      *gorm.DB
	stt     SpeechToText
	storage BlobStorage
}

// Transcribe will continuously poll the database for new messages and run
// speech-to-text on them.
func Transcribe(ctx context.Context, logger *zap.Logger, db *gorm.DB, stt SpeechToText, storage BlobStorage) error {
	t := Transcriber{
		logger:  logger,
		db:      db.WithContext(ctx),
		stt:     stt,
		storage: storage,
	}

	// Note: there's no point polling more rapidly than chunks are generated
	ticker := time.NewTicker(ChunkLength)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := t.transcribe(ctx); err != nil {
				return err
			}
		}
	}
}

func (t *Transcriber) transcribe(ctx context.Context) error {
	transmissions, err := untranscribedTransmissions(t.db)
	if err != nil {
		return err
	} else if len(transmissions) == 0 {
		// Nothing to do...
		return nil
	}

	t.logger.Debug("Transcribing", zap.Any("transmissions", transmissions))

	var urls []*url.URL

	for _, transmission := range transmissions {
		key, err := ParseBlobKey(transmission.Sha256)
		if err != nil {
			return fmt.Errorf("unable to parse %q as a blob key: %w", transmission.Sha256, err)
		}
		url, err := t.storage.Link(ctx, key)
		if err != nil {
			return fmt.Errorf("unable to get a link to %q: %w", key, err)
		}
		urls = append(urls, url)
	}

	transcriptions, err := t.stt.SpeechToText(ctx, urls)
	if err != nil {
		return fmt.Errorf("transcription failed: %w", err)
	} else if len(transcriptions) != len(urls) {
		return fmt.Errorf("transcriber returned %d strings, but expected %d", len(transcriptions), len(urls))
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
		return fmt.Errorf("unable to save the new transcriptions: %w", err)
	}

	t.logger.Info("Saved transcriptions", zap.Any("transcriptions", models))

	return nil
}

// untranscribedTransmissions will query the database for all Transmissions that
// haven't yet been transcribed.
func untranscribedTransmissions(db *gorm.DB) ([]Transmission, error) {
	var transmissions []Transmission
	err := db.Joins("LEFT JOIN transcriptions ON transcriptions.transmission_id = transmissions.id").
		Where("transcriptions.id IS NULL").
		Limit(MaxSpeechToTextBatchSize).
		Find(&transmissions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query untranscribed transmissions: %w", err)
	}

	return transmissions, nil
}
