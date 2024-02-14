package radiochatter

import (
	"context"
	"path"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	cb := archiveCallbacks(ctx, ch, dummyNow)
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
			{
				Path:      path.Join(temp, "chunk_0.mp3"),
				Timestamp: timestamp(0),
				Pieces: []audioSpan{
					{Start: 18323800000, End: 22560400000},
					{Start: 26447300000, End: 28632600000},
					{Start: 29799600000, End: 58637000000},
				},
			},
			{
				Path:      path.Join(temp, "chunk_1.mp3"),
				Timestamp: timestamp(60 * time.Second),
				Pieces: []audioSpan{
					{Start: 1848000000, End: 9876800000},
					{Start: 11608400000, End: 19490900000},
					{Start: 22476299999, End: 24196100000},
					{Start: 26355800000, End: 31028599999},
					{Start: 32477800000, End: 32998500000},
					{Start: 34763400000, End: 40691000000},
					{Start: 41918000000, End: 60000000000},
				},
			},
			{
				Path:      path.Join(temp, "chunk_2.mp3"),
				Timestamp: timestamp(120 * time.Second),
				Pieces: []audioSpan{
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
	ctx := testContext(t)
	cb := archiveCallbacks(ctx, ch, dummyNow)

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
			{
				Path:      "output000.mp3",
				Timestamp: timestamp(0),
				Pieces: []audioSpan{
					{Start: 19029900000, End: 24462600000},
					{Start: 31306100000, End: 36254100000},
					{Start: 60418600000, End: 60000000000},
				},
			},
			{
				Path:      "output001.mp3",
				Timestamp: timestamp(60 * time.Second),
				Pieces: []audioSpan{
					{0, 5108099999},
					{36096400000, 40403000000},
					{42443000000, 50320000000},
					{52502000000, 58398000000},
				},
			},
			{
				Path:      "output002.mp3",
				Timestamp: timestamp(120 * time.Second),
			},
		},
		ops,
	)
}

func TestJustSilence(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	ctx := testContext(t)
	cb := archiveCallbacks(ctx, ch, dummyNow)

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
			{
				Path:      "chunk_0.mp3",
				Timestamp: timestamp(0),
			},
			{
				Path:      "chunk_1.mp3",
				Timestamp: timestamp(60 * time.Second),
			},
		},
		ops,
	)
}

func TestClipContainingAudio(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	ctx := testContext(t)
	cb := archiveCallbacks(ctx, ch, dummyNow)

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
			{
				Path:      "chunk_0.mp3",
				Timestamp: timestamp(0),
				Pieces: []audioSpan{
					{Start: 10 * time.Second, End: 15 * time.Second},
				},
			},
		},
		ops,
	)
}

func TestAudioInSecondClip(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	ctx := testContext(t)
	cb := archiveCallbacks(ctx, ch, dummyNow)

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
			{
				Path:      "chunk_0.mp3",
				Timestamp: timestamp(0),
			},
			{
				Path:      "chunk_1.mp3",
				Timestamp: timestamp(60 * time.Second),
				Pieces: []audioSpan{
					{Start: 5 * time.Second, End: 10 * time.Second},
				},
			},
		},
		ops,
	)
}

func TestAudioAcrossChunkBoundary(t *testing.T) {
	ch := make(chan ArchiveOperation, 16)
	ctx := testContext(t)
	cb := archiveCallbacks(ctx, ch, dummyNow)

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
			{
				Path:      "chunk_0.mp3",
				Timestamp: timestamp(0),
				Pieces: []audioSpan{
					{Start: 50 * time.Second, End: 60 * time.Second},
				},
			},
			{
				Path:      "chunk_1.mp3",
				Timestamp: timestamp(60 * time.Second),
				Pieces: []audioSpan{
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

func TestSplitAndArchiveRecording(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	logger := zaptest.NewLogger(t)
	ctx := testContext(t)
	group, ctx := errgroup.WithContext(ctx)
	db := testDatabase(ctx, t)
	stream := Stream{DisplayName: "Test", Url: testRecording(t)}
	assert.NoError(t, db.Save(&stream).Error)
	storage := NewOnDiskStorage(logger, t.TempDir())

	archiveOps := make(chan ArchiveOperation)
	temp, cleanup := mkdtemp(logger)
	defer cleanup()
	group.Go(preprocess(ctx, logger.Named("preprocess"), stream.Url, temp, archiveOps))
	group.Go(archive(ctx, logger.Named("archive"), archiveOps, storage, db, stream))
	assert.NoError(t, group.Wait())

	assert.NoError(t, db.Preload("Chunks").Preload("Chunks.Transmissions").Find(&stream).Error)
	assert.Equal(t, 3, len(stream.Chunks))
	transmisions := stream.Chunks[1].Transmissions
	assert.Equal(t, 7, len(transmisions))
	sort.Slice(transmisions, func(i, j int) bool { return transmisions[i].TimeStamp.Before(transmisions[j].TimeStamp) })
	transmission := stream.Chunks[1].Transmissions[0]
	assert.Equal(t, "7d865c69589b323c95dcb2c5aab008559efcfe226284aa8a70db5d0c01f04e71", transmission.Sha256)
	assert.Equal(t, stream.Chunks[1].ID, transmission.ChunkID)
	assert.Equal(t, 8028800*time.Microsecond, transmission.Length)
}

func testDatabase(ctx context.Context, t *testing.T) *gorm.DB {
	t.Helper()

	filename := path.Join(t.TempDir(), "db.sqlite3")

	db, err := gorm.Open(sqlite.Open(filename))
	assert.NoError(t, err)

	assert.NoError(t, Migrate(ctx, db))

	return db
}
