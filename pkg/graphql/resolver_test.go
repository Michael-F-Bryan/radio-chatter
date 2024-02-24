package graphql

import (
	"context"
	"path"
	"testing"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/middleware"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/on_disk_storage"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

	return middleware.WithLogger(context.Background(), zaptest.NewLogger(t))
}
