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
			saveChunk{path: path.Join(temp, "chunk_0.mp3")},
			splitAudioSnippets{
				path: path.Join(temp, "chunk_0.mp3"),
				pieces: []audioSpan{
					{start: 18323800000, end: 22560400000},
					{start: 26447300000, end: 28632600000},
					{start: 29799600000, end: 58637000000},
				},
			},
			saveChunk{path: path.Join(temp, "chunk_1.mp3")},
			splitAudioSnippets{
				path: path.Join(temp, "chunk_1.mp3"),
				pieces: []audioSpan{
					{start: 1848000000, end: 9876800000},
					{start: 11608400000, end: 19490900000},
					{start: 22476299999, end: 24196100000},
					{start: 26355800000, end: 31028599999},
					{start: 32477800000, end: 32998500000},
					{start: 34763400000, end: 40691000000},
					{start: 41918000000, end: 60000000000},
				},
			},
			saveChunk{path: path.Join(temp, "chunk_2.mp3")},
			splitAudioSnippets{
				path: path.Join(temp, "chunk_2.mp3"),
				pieces: []audioSpan{
					{start: 0, end: 18446000000},
					{start: 21415000000, end: 23333000000},
				},
			},
		},
		ops,
	)
}

func TestReplayStderrToArchiver(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ch := make(chan ArchiveOperation, 16)
	cb := ArchiveCallbacks(ch)

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
			saveChunk{path: "output000.mp3"},
			splitAudioSnippets{
				path: "output000.mp3",
				pieces: []audioSpan{
					{start: 19029900000, end: 24462600000},
					{start: 31306100000, end: 36254100000},
					{start: 60418600000, end: 60000000000},
				},
			},
			saveChunk{path: "output001.mp3"},
			splitAudioSnippets{
				path: "output001.mp3",
				pieces: []audioSpan{
					{0, 5108099999},
					{36096400000, 40403000000},
					{42443000000, 50320000000},
					{52502000000, 58398000000},
				},
			},
			saveChunk{path: "output002.mp3"},
		},
		ops,
	)
}

func TestJustSilence(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	cb := ArchiveCallbacks(ch)

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
			saveChunk{path: "chunk_0.mp3"},
			saveChunk{path: "chunk_1.mp3"},
		},
		ops,
	)
}

func TestClipContainingAudio(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	cb := ArchiveCallbacks(ch)

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
			saveChunk{path: "chunk_0.mp3"},
			splitAudioSnippets{
				path: "chunk_0.mp3",
				pieces: []audioSpan{
					{start: 10 * time.Second, end: 15 * time.Second},
				},
			},
		},
		ops,
	)
}

func TestAudioInSecondClip(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	cb := ArchiveCallbacks(ch)

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
			saveChunk{path: "chunk_0.mp3"},
			saveChunk{path: "chunk_1.mp3"},
			splitAudioSnippets{
				path: "chunk_1.mp3",
				pieces: []audioSpan{
					{start: 5 * time.Second, end: 10 * time.Second},
				},
			},
		},
		ops,
	)
}

func TestAudioAcrossChunkBoundary(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	cb := ArchiveCallbacks(ch)

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
			saveChunk{path: "chunk_0.mp3"},
			splitAudioSnippets{
				path: "chunk_0.mp3",
				pieces: []audioSpan{
					{start: 50 * time.Second, end: 60 * time.Second},
				},
			},
			saveChunk{path: "chunk_1.mp3"},
			splitAudioSnippets{
				path: "chunk_1.mp3",
				pieces: []audioSpan{
					{start: 0 * time.Second, end: 5 * time.Second},
				},
			},
		},
		ops,
	)
}
