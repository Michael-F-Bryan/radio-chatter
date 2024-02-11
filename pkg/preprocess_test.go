package radiochatter

import (
	"context"
	_ "embed"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

//go:embed stderr.txt
var stderr string

//go:embed recording.mp3
var recording []byte

type endPair struct {
	end      time.Duration
	duration time.Duration
}

func TestParseStderr(t *testing.T) {
	logger := zaptest.NewLogger(t)
	reader := strings.NewReader(stderr)
	var cb eventData

	err := parseStderr(logger, reader, cb.Callbacks(t))

	assert.NoError(t, err)
	assert.True(t, cb.started)
	assert.Equal(t, []string{"output000.mp3", "output001.mp3", "output002.mp3"}, cb.writing)
	assert.Equal(t, []time.Duration{0, 24462600000, 36254100000, 65108099999, 100403000000, 110320000000, 118398000000}, cb.silenceStart)
	assert.Equal(
		t,
		[]endPair{
			{19029900000, 19029900000},
			{31306100000, 6843500000},
			{60418600000, 24164500000},
			{96096400000, 30988300000},
			{102443000000, 2040620000},
			{112502000000, 2181379999},
			{120913000000, 2515120000},
		},
		cb.silenceEnd,
	)
	assert.Empty(t, cb.unknown)
}

func TestRealRecording(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	logger := zaptest.NewLogger(t)
	ctx := testContext(t)
	input := testRecording(t)
	temp := t.TempDir()
	var e eventData

	err := Preprocess(ctx, logger, input, temp, e.Callbacks(t))

	assert.NoError(t, err)
	assert.Equal(t,
		eventData{
			started: true,
			writing: []string{
				path.Join(temp, "chunk_0.mp3"),
				path.Join(temp, "chunk_1.mp3"),
				path.Join(temp, "chunk_2.mp3"),
			},
			silenceStart: []time.Duration{
				0,
				22560400000,
				28632600000,
				58637000000,
				69876800000,
				79490900000,
				84196100000,
				91028599999,
				92998500000,
				100691000000,
				138446000000,
				143333000000,
			},
			silenceEnd: []endPair{
				{18323800000, 18323800000},
				{26447300000, 3886880000},
				{29799600000, 1167000000},
				{61848000000, 3211000000},
				{71608400000, 1731620000},
				{82476299999, 2985370000},
				{86355800000, 2159630000},
				{92477800000, 1449120000},
				{94763400000, 1764870000},
				{101918000000, 1227379999},
				{141415000000, 2968370000},
				{158256000000, 14923199999},
			},
			unknown: nil,
		},
		e,
	)
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel := context.WithDeadline(context.Background(), deadline)
		t.Cleanup(cancel)
		return ctx
	}

	return context.Background()
}

type eventData struct {
	started      bool
	writing      []string
	silenceStart []time.Duration
	silenceEnd   []endPair
	unknown      []ComponentMessage
}

func (e *eventData) Callbacks(t *testing.T) PreprocessingCallbacks {
	t.Helper()

	return PreprocessingCallbacks{
		DownloadStarted: func() { e.started = true },
		StartWriting: func(path string) {
			e.writing = append(e.writing, path)
		},
		SilenceStart: func(t time.Duration) {
			e.silenceStart = append(e.silenceStart, t)
		},
		UnknownMessage: func(msg ComponentMessage) {
			t.Errorf("Unknown message: %#v", msg)
		},
		SilenceEnd: func(t, duration time.Duration) {
			e.silenceEnd = append(e.silenceEnd, endPair{t, duration})
		},
	}
}

func testRecording(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	filename := path.Join(dir, "recording.mp3")

	err := os.WriteFile(filename, recording, 0766)
	assert.NoError(t, err)

	return filename
}
