package radiochatter

import (
	"net/url"
	"os/exec"
	"testing"
	"time"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/on_disk_storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"
)

func TestFindUntranscribedTransmissions(t *testing.T) {
	ctx := testContext(t)
	db := testDatabase(ctx, t)
	stream := Stream{DisplayName: "Test", Url: "..."}
	assert.NoError(t, db.Save(&stream).Error)
	chunk := Chunk{StreamID: stream.ID}
	assert.NoError(t, db.Save(&chunk).Error)
	transmission := Transmission{ChunkID: chunk.ID}
	assert.NoError(t, db.Save(&transmission).Error)

	untranscribed, err := untranscribedTransmissions(db, 1000)

	assert.NoError(t, err)
	assert.Len(t, untranscribed, 1)
	untranscribed[0].Model = gorm.Model{ID: untranscribed[0].ID}
	transmission.Model = gorm.Model{ID: transmission.ID}
	assert.Equal(t, []Transmission{transmission}, untranscribed)
}

func TestTranscribeUsingWhisper(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	requires(t, whisperCommand, ffmpegCommand)

	ctx := testContext(t)
	logger := zaptest.NewLogger(t)
	storage, err := on_disk_storage.New(logger, t.TempDir())
	assert.NoError(t, err)
	defer storage.Close()
	db := testDatabase(ctx, t)
	recording := testRecording(t)
	state := ArchiveState{
		Logger:  logger,
		Storage: storage,
		DB:      db,
	}
	span := audioSpan{Start: 18323800000, End: 22560400000}
	transmission, err := splitAudio(ctx, state, recording, span, Chunk{})
	assert.NoError(t, err)
	w := NewWhisperTranscriber(logger)
	key, err := blob.ParseKey(transmission.Sha256)
	assert.NoError(t, err)
	recordingURL, err := storage.Link(ctx, key, 1*time.Hour)
	assert.NoError(t, err)

	transcriptions, err := w.SpeechToText(ctx, []*url.URL{recordingURL})

	assert.NoError(t, err)
	assert.Equal(t, []string{"Okay, out to Verock, over.\n"}, transcriptions)
}

func requires(t *testing.T, programs ...string) {
	for _, program := range programs {
		if _, err := exec.LookPath(program); err != nil {
			t.Skipf("%q isn't installed: %e", program, err)
		}
	}
}
