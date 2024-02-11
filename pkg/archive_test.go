package radiochatter

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestArchive60SecondChunks(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	logger := zaptest.NewLogger(t)
	ctx := testContext(t)
	input := testRecording(t)
	temp := t.TempDir()
	ch := make(chan ArchiveOperation, 16)
	cb := ArchiveCallbacks(ch)
	go func() {
		defer close(ch)
		err := Preprocess(ctx, logger, input, temp, cb)
		assert.NoError(t, err)
	}()

	var ops []ArchiveOperation
	for op := range ch {
		ops = append(ops, op)
	}

	assert.Equal(
		t,
		[]ArchiveOperation{
			saveChunk{path: path.Join(temp, "output0.mp3")},
			saveChunk{path: path.Join(temp, "output1.mp3")},
			saveChunk{path: path.Join(temp, "output2.mp3")},
		},
		ops,
	)
}
