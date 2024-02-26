package graphql

import (
	"context"
	"log"
	"os"
	"path"
	"testing"
	"time"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/middleware"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/on_disk_storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

func TestGetStreamByID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := testContext(t)
	storage, err := on_disk_storage.New(logger, t.TempDir())
	assert.NoError(t, err)
	defer storage.Close()
	resolver := Resolver{
		DB:      testDatabase(ctx, t),
		Storage: storage,
	}
	stream := radiochatter.Stream{DisplayName: "Test", Url: "..."}
	assert.NoError(t, resolver.DB.Save(&stream).Error)
	id := modelId(stream)

	got, err := resolver.Query().GetStreamByID(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, streamToGraphQL(stream), *got)
}

func TestGetChunkByID(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := testContext(t)
	storage, err := on_disk_storage.New(logger, t.TempDir())
	assert.NoError(t, err)
	defer storage.Close()
	resolver := Resolver{
		DB:      testDatabase(ctx, t),
		Storage: storage,
	}
	stream := radiochatter.Stream{DisplayName: "Test", Url: "..."}
	assert.NoError(t, resolver.DB.Save(&stream).Error)
	chunk := radiochatter.Chunk{Sha256: "asdf", StreamID: stream.ID}
	assert.NoError(t, resolver.DB.Save(&chunk).Error)
	id := modelId(chunk)

	got, err := resolver.Query().GetChunkByID(ctx, id)

	assert.NoError(t, err)
	assert.Equal(t, chunkToGraphQL(chunk), *got)
}

func TestSubscribeToNewChunks(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx, cancel := context.WithCancel(testContext(t))
	defer cancel()
	db := testDatabase(ctx, t)
	storage, err := on_disk_storage.New(logger, t.TempDir())
	assert.NoError(t, err)
	defer storage.Close()
	resolver := Resolver{
		DB:           db,
		Storage:      storage,
		PollInterval: 20 * time.Millisecond,
	}
	stream := radiochatter.Stream{
		DisplayName: "Test",
	}
	assert.NoError(t, db.Save(&stream).Error)

	// First we subscribe to receive new chunks
	ch, err := resolver.Subscription().Chunks(ctx)
	assert.NoError(t, err)
	// Wait for the poller to get started
	time.Sleep(2 * resolver.PollInterval)
	// Pretend something saved a chunk in the background
	firstChunk := radiochatter.Chunk{
		Sha256:   "first",
		StreamID: stream.ID,
	}
	assert.NoError(t, db.Save(&firstChunk).Error)
	// The subscription should receive it
	value, ok := <-ch
	assert.True(t, ok)
	assert.Equal(t, chunkToGraphQL(firstChunk), *value)
	// Now we'll cancel the subscription
	cancel()
	// The channel should be closed
	value, ok = <-ch
	assert.False(t, ok)
	assert.Zero(t, value)
	// To be sure, let's save another chunk
	secondChunk := radiochatter.Chunk{
		Sha256:   "second",
		StreamID: stream.ID,
	}
	assert.NoError(t, db.Save(&secondChunk).Error)
	// And the channel is still closed
	assert.Zero(t, <-ch)
}

func TestSubscribeToTransmissionsForStream(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx, cancel := context.WithCancel(testContext(t))
	defer cancel()
	db := testDatabase(ctx, t)
	db = db.Session(&gorm.Session{
		Logger: glog.New(log.New(os.Stderr, "[SQL] ", log.Flags()), glog.Config{LogLevel: glog.Info}),
	})
	storage, err := on_disk_storage.New(logger, t.TempDir())
	assert.NoError(t, err)
	defer storage.Close()
	resolver := Resolver{
		DB:           db,
		Storage:      storage,
		PollInterval: 20 * time.Millisecond,
	}
	stream := radiochatter.Stream{DisplayName: "Test"}
	assert.NoError(t, db.Save(&stream).Error)
	chunk := radiochatter.Chunk{
		Sha256:   "test-chunk",
		StreamID: stream.ID,
	}
	assert.NoError(t, db.Save(&chunk).Error)

	// First we subscribe to receive new transmissions
	ch, err := resolver.Subscription().Transmissions(ctx, modelId(&stream))
	assert.NoError(t, err)
	// Wait for the poller to get started
	time.Sleep(2 * resolver.PollInterval)
	t.Log("First")
	// Pretend something saved a chunk in the background
	firstTransmission := radiochatter.Transmission{
		Sha256:  "first",
		ChunkID: chunk.ID,
	}
	assert.NoError(t, db.Save(&firstTransmission).Error)
	t.Log("Second")
	// The subscription should receive it
	value, ok := <-ch
	assert.True(t, ok)
	assert.Equal(t, transmissionToGraphQL(firstTransmission), *value)
	t.Log("Third")
	// We save to a different stream and make sure we don't get anything
	otherStream := radiochatter.Stream{DisplayName: "Other"}
	assert.NoError(t, db.Save(&otherStream).Error)
	otherChunk := radiochatter.Chunk{StreamID: otherStream.ID}
	assert.NoError(t, db.Save(&otherChunk).Error)
	dummyTransmission := radiochatter.Transmission{ChunkID: otherChunk.ID}
	assert.NoError(t, db.Save(&dummyTransmission).Error)
	t.Log("Fourth")
	select {
	case got := <-ch:
		t.Fatalf("Shouldn't have received anything but got %v", got)
	default:
		// The channel shouldn't have anything - success
	}
	// Now we'll cancel the subscription
	cancel()
	// The channel should be closed
	value, ok = <-ch
	assert.False(t, ok)
	assert.Zero(t, value)
}

func testDatabase(ctx context.Context, t *testing.T) *gorm.DB {
	t.Helper()

	filename := path.Join(t.TempDir(), "db.sqlite3")

	db, err := gorm.Open(sqlite.Open(filename))
	assert.NoError(t, err)

	assert.NoError(t, radiochatter.Migrate(ctx, db))

	return db
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		t.Cleanup(cancel)
		return ctx
	}

	logger := zaptest.NewLogger(t)
	t.Cleanup(func() { _ = logger.Sync() })
	return middleware.WithLogger(context.Background(), logger)
}
