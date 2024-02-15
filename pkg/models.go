package radiochatter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Stream is an audio stream downloaded from the internet.
type Stream struct {
	gorm.Model
	// The human-friendly name for this stream.
	DisplayName string `gorm:"unique"`
	// A URL that can be passed to ffmpeg to download the stream.
	Url string
	// Downloaded chunks.
	Chunks []Chunk `gorm:"constraint:OnDelete:CASCADE"`
}

// Chunk is a raw chunk of audio downloaded from a particular stream.
type Chunk struct {
	gorm.Model
	// When the audio was produced.
	TimeStamp time.Time
	// A hex-encoded hash of the audio clip.
	Sha256 string
	// The stream this clip belongs to.
	StreamID uint
	// Messages that were transmitted in this chunk.
	Transmissions []Transmission `gorm:"constraint:OnDelete:CASCADE"`
}

// Transmission contains a single radio transmission.
type Transmission struct {
	gorm.Model
	// When the transmission was made.
	TimeStamp time.Time
	// How long the transmission goes for.
	Length time.Duration
	// A hex-encoded hash of the audio clip.
	Sha256 string
	// The chunk this transmission came from.
	ChunkID       uint
	Transcription *Transcription `gorm:"constraint:OnDelete:CASCADE"`
}

// Transcription is the result of running speech-to-text on a Transmission.
type Transcription struct {
	gorm.Model
	TransmissionID uint
	// The content of the transmission.
	Content string
}

// Migrate will apply any necessary migrations to the database.
func Migrate(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).AutoMigrate(&Stream{}, &Chunk{}, &Transmission{}, &Transcription{})
}

var databaseOpeners = map[string]func(string) gorm.Dialector{
	"sqlite3":  sqlite.Open,
	"postgres": postgres.Open,
}

func OpenDatabase(ctx context.Context, logger *zap.Logger, dbDriver string, conn string) (*gorm.DB, error) {
	opener, ok := databaseOpeners[dbDriver]
	if !ok {
		var dbTypes []string
		for k := range databaseOpeners {
			dbTypes = append(dbTypes, k)
		}
		return nil, fmt.Errorf("unknown database driver, %q, expected one of %s", dbDriver, strings.Join(dbTypes, ", "))
	}

	db, err := gorm.Open(opener(conn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	logger.Debug(
		"Applying migrations",
		zap.String("database", dbDriver),
		zap.String("connection-string", conn),
	)

	if err := Migrate(ctx, db); err != nil {
		return nil, err
	}

	return db, nil
}
