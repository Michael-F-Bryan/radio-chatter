package on_disk_storage

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestStoreBlobAndReadItBack(t *testing.T) {
	logger := zaptest.NewLogger(t)
	storage, err := New(logger, t.TempDir())
	assert.NoError(t, err)
	defer storage.Close()
	blob := "Hello, World"

	key, err := storage.Store(context.Background(), []byte(blob))
	assert.NoError(t, err)
	link, err := storage.Link(context.Background(), key, 1*time.Hour)
	assert.NoError(t, err)
	response, err := http.Get(link.String())
	assert.NoError(t, err)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	assert.NoError(t, err)

	assert.Equal(t, blob, string(body))
}
