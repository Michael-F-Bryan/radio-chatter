package radiochatter

import (
	_ "embed"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

//go:embed stderr.txt
var stderr string

type endPair struct {
	end      time.Duration
	duration time.Duration
}

func TestParseStderr(t *testing.T) {
	logger := zaptest.NewLogger(t)
	reader := strings.NewReader(stderr)
	started := false
	writing := []string{}
	silenceStart := []time.Duration{}
	silenceEnd := []endPair{}
	cb := FfmpegCallbacks{
		DownloadStarted: func() { started = true },
		StartWriting: func(path string) {
			writing = append(writing, path)
		},
		SilenceStart: func(t time.Duration) {
			silenceStart = append(silenceStart, t)
		},
		UnknownMessage: func(msg ComponentMessage) {
			t.Errorf("Unknown message: %#v", msg)
		},
		SilenceEnd: func(t, duration time.Duration) {
			silenceEnd = append(silenceEnd, endPair{t, duration})
		},
	}

	parseStderr(logger, reader, cb)

	assert.True(t, started)
	assert.Equal(t, []string{"output000.mp3", "output001.mp3", "output002.mp3"}, writing)
	assert.Equal(t, []time.Duration{0, 24462600000, 36254100000, 65108099999, 100403000000, 110320000000, 118398000000}, silenceStart)
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
		silenceEnd,
	)
}
