package radiochatter

import (
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestArchiveRealRecording(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	logger := zaptest.NewLogger(t)
	ctx := testContext(t)
	input := testRecording(t)
	temp := t.TempDir()
	ch := make(chan ArchiveOperation)
	cb := archiveCallbacks(ch, dummyNow)
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
			saveChunk{
				path:      path.Join(temp, "chunk_0.mp3"),
				timestamp: timestamp(0),
			},
			splitAudioSnippets{
				path:      path.Join(temp, "chunk_0.mp3"),
				timestamp: timestamp(0),
				pieces: []audioSpan{
					{Start: 18323800000, End: 22560400000},
					{Start: 26447300000, End: 28632600000},
					{Start: 29799600000, End: 58637000000},
				},
			},
			saveChunk{
				path:      path.Join(temp, "chunk_1.mp3"),
				timestamp: timestamp(60 * time.Second),
			},
			splitAudioSnippets{
				path:      path.Join(temp, "chunk_1.mp3"),
				timestamp: timestamp(60 * time.Second),
				pieces: []audioSpan{
					{Start: 1848000000, End: 9876800000},
					{Start: 11608400000, End: 19490900000},
					{Start: 22476299999, End: 24196100000},
					{Start: 26355800000, End: 31028599999},
					{Start: 32477800000, End: 32998500000},
					{Start: 34763400000, End: 40691000000},
					{Start: 41918000000, End: 60000000000},
				},
			},
			saveChunk{
				path:      path.Join(temp, "chunk_2.mp3"),
				timestamp: timestamp(120 * time.Second),
			},
			splitAudioSnippets{
				path:      path.Join(temp, "chunk_2.mp3"),
				timestamp: timestamp(120 * time.Second),
				pieces: []audioSpan{
					{Start: 0, End: 18446000000},
					{Start: 21415000000, End: 23333000000},
				},
			},
		},
		ops,
	)
}

func TestReplayStderrToArchiver(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ch := make(chan ArchiveOperation, 16)
	cb := archiveCallbacks(ch, dummyNow)

	err := parseStderr(logger, strings.NewReader(stderr), cb)

	close(ch)
	var ops []ArchiveOperation
	for op := range ch {
		ops = append(ops, op)
	}

	assert.NoError(t, err)
	assert.Equal(
		t,
		[]ArchiveOperation{
			saveChunk{
				path:      "output000.mp3",
				timestamp: timestamp(0),
			},
			splitAudioSnippets{
				path:      "output000.mp3",
				timestamp: timestamp(0),
				pieces: []audioSpan{
					{Start: 19029900000, End: 24462600000},
					{Start: 31306100000, End: 36254100000},
					{Start: 60418600000, End: 60000000000},
				},
			},
			saveChunk{
				path:      "output001.mp3",
				timestamp: timestamp(60 * time.Second),
			},
			splitAudioSnippets{
				path:      "output001.mp3",
				timestamp: timestamp(60 * time.Second),
				pieces: []audioSpan{
					{0, 5108099999},
					{36096400000, 40403000000},
					{42443000000, 50320000000},
					{52502000000, 58398000000},
				},
			},
			saveChunk{
				path:      "output002.mp3",
				timestamp: timestamp(120 * time.Second),
			},
		},
		ops,
	)
}

func TestJustSilence(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	cb := archiveCallbacks(ch, dummyNow)

	// First we start downloading
	cb.onDownloadStarted()
	// and open the first chunk
	cb.onStartWriting("chunk_0.mp3")
	// it starts with silence
	cb.onSilenceStart(0)
	// We start saving to our second chunk
	cb.onStartWriting("chunk_1.mp3")
	// End of input
	cb.onSilenceEnd(100*time.Second, 100*time.Second)
	cb.onFinished()
	close(ch)

	var ops []ArchiveOperation
	for op := range ch {
		ops = append(ops, op)
	}

	assert.Equal(
		t,
		[]ArchiveOperation{
			saveChunk{
				path:      "chunk_0.mp3",
				timestamp: timestamp(0),
			},
			saveChunk{
				path:      "chunk_1.mp3",
				timestamp: timestamp(60 * time.Second),
			},
		},
		ops,
	)
}

func TestClipContainingAudio(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	cb := archiveCallbacks(ch, dummyNow)

	// First we start downloading
	cb.onDownloadStarted()
	// and open the first chunk
	cb.onStartWriting("chunk_0.mp3")
	// it starts with silence
	cb.onSilenceStart(0)
	// but someone starts talking after 10 seconds
	cb.onSilenceEnd(10*time.Second, 10*time.Second)
	// They finish saying their bit after 5 seconds, then it's silence again
	cb.onSilenceStart(15 * time.Second)
	// And finally, we reach the end of input
	cb.onSilenceEnd(60*time.Second, 45*time.Second)
	cb.onFinished()
	close(ch)

	var ops []ArchiveOperation
	for op := range ch {
		ops = append(ops, op)
	}

	assert.Equal(
		t,
		[]ArchiveOperation{
			saveChunk{
				path:      "chunk_0.mp3",
				timestamp: timestamp(0),
			},
			splitAudioSnippets{
				path:      "chunk_0.mp3",
				timestamp: timestamp(0),
				pieces: []audioSpan{
					{Start: 10 * time.Second, End: 15 * time.Second},
				},
			},
		},
		ops,
	)
}

func TestAudioInSecondClip(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	cb := archiveCallbacks(ch, dummyNow)

	// First we start downloading
	cb.onDownloadStarted()
	// and open the first chunk
	cb.onStartWriting("chunk_0.mp3")
	// it starts with silence
	cb.onSilenceStart(0)
	// We start the next chunk
	cb.onStartWriting("chunk_1.mp3")
	// Audio starts 5 seconds into the next 60-second chunk
	cb.onSilenceEnd(65*time.Second, 65*time.Second)
	// They finish saying their bit after 5 seconds, then it's silence again
	cb.onSilenceStart(70 * time.Second)
	// And finally, we reach the end of input
	cb.onSilenceEnd(120*time.Second, 50*time.Second)
	cb.onFinished()
	close(ch)

	var ops []ArchiveOperation
	for op := range ch {
		ops = append(ops, op)
	}

	assert.Equal(
		t,
		[]ArchiveOperation{
			saveChunk{
				path:      "chunk_0.mp3",
				timestamp: timestamp(0),
			},
			saveChunk{
				path:      "chunk_1.mp3",
				timestamp: timestamp(60 * time.Second),
			},
			splitAudioSnippets{
				path:      "chunk_1.mp3",
				timestamp: timestamp(60 * time.Second),
				pieces: []audioSpan{
					{Start: 5 * time.Second, End: 10 * time.Second},
				},
			},
		},
		ops,
	)
}

func TestAudioAcrossChunkBoundary(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	cb := archiveCallbacks(ch, dummyNow)

	// First we start downloading
	cb.onDownloadStarted()
	// and open the first chunk
	cb.onStartWriting("chunk_0.mp3")
	// it starts with silence
	cb.onSilenceStart(0)
	// but someone starts talking after 50 seconds
	cb.onSilenceEnd(50*time.Second, 50*time.Second)
	// The next chunk starts before they finished talking
	cb.onStartWriting("chunk_1.mp3")
	// And they finish talking about 5 seconds in
	cb.onSilenceStart(65 * time.Second)
	// And finally, we reach the end of input
	cb.onSilenceEnd(120*time.Second, 55*time.Second)
	cb.onFinished()
	close(ch)

	var ops []ArchiveOperation
	for op := range ch {
		ops = append(ops, op)
	}

	assert.Equal(
		t,
		[]ArchiveOperation{
			saveChunk{
				path:      "chunk_0.mp3",
				timestamp: timestamp(0),
			},
			splitAudioSnippets{
				path:      "chunk_0.mp3",
				timestamp: timestamp(0),
				pieces: []audioSpan{
					{Start: 50 * time.Second, End: 60 * time.Second},
				},
			},
			saveChunk{
				path:      "chunk_1.mp3",
				timestamp: timestamp(60 * time.Second),
			},
			splitAudioSnippets{
				path:      "chunk_1.mp3",
				timestamp: timestamp(60 * time.Second),
				pieces: []audioSpan{
					{Start: 0 * time.Second, End: 5 * time.Second},
				},
			},
		},
		ops,
	)
}

// dummyNow returns a stable timestamp that can be used instead of relying on
// time.Now().
func dummyNow() time.Time {
	return time.Time{}
}

// timestamp returns a time.Time that is a certain amount of time after
// dummyNow().
func timestamp(d time.Duration) time.Time {
	t := time.Time{}
	return t.Add(d).UTC()
}
