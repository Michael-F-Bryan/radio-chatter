package radiochatter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const defaultGracefulShutdown = 10 * time.Second
const defaultCommand = "ffmpeg"
const FeedUrl = "https://broadcastify.cdnstream1.com/39131"

// messagePattern will match a string like "[silencedetect @ 0x600000c583c0] ..."
var messagePattern = regexp.MustCompile(`^\[(\S+) @ [\d\w]+\]\s*(.*)$`)

// openingFilePattern matches a string like "Opening '/path/to/file.mp3' for writing"
var openingFilePattern = regexp.MustCompile(`^Opening '([^']+)' for writing$`)

var silenceStartPattern = regexp.MustCompile(`silence_start: (\d+(?:\.\d+)?)`)
var silenceEndPattern = regexp.MustCompile(`silence_end: (\d+(?:\.\d+)?) \| silence_duration: (\d+(?:\.\d+)?)`)

func RunFfmpeg(ctx context.Context, logger *zap.Logger, input string, outputDir string, cb FfmpegCallbacks) error {
	args := []string{
		"-i", input,
		// Use a filter to detect silence and print its timestamps
		"-af", "silencedetect=noise=-30dB:d=1",
		// Split into 60-second chunks
		"-f", "segment", "-segment_time", "60",
		// Clean up stderr so it's easier to parse
		"-hide_banner", "-nostdin", "-nostats",
		// the output path
		path.Join(outputDir, "output%d.mp3"),
	}

	cmd := exec.CommandContext(ctx, defaultCommand, args...)
	cmd.Stdout = os.Stdout

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	go parseStderr(logger, stderr, cb)

	// Note: We want to give ffmpeg a chance to flush its buffers and shut down
	// gracefully, so when the context is cancelled we'll first send a SIGINT,
	// then wait a bit to let the command exit. If it doesn't exit in time, it
	// will be forcefully killed.
	cmd.WaitDelay = defaultGracefulShutdown
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unable to start %q: %w", cmd, err)
	}

	logger.Debug(
		"Started ffmpeg",
		zap.Stringer("cmd", cmd),
		zap.Int("pid", cmd.Process.Pid),
	)

	err = cmd.Wait()

	// As an edge case, ffmpeg will exit with a 255 exit code when you send it a
	// SIGINT (e.g. because we told it to shut down). This is an expected
	// condition, so don't treat it like an error.
	if cmd.ProcessState.ExitCode() == 255 {
		return nil
	}

	return err
}

// parseStderr reads the output from ffmpeg and triggers callbacks to notify
// the caller when certain events occur.
func parseStderr(logger *zap.Logger, stderr io.Reader, cb FfmpegCallbacks) {
	state := state{
		running: false,
		cb:      cb,
		logger:  logger,
	}
	buffer := bufio.NewScanner(stderr)

	for buffer.Scan() {
		line := buffer.Text()

		match := messagePattern.FindStringSubmatch(line)
		if match != nil {
			msg := ComponentMessage{
				Timestamp: time.Now(),
				Component: match[1],
				Payload:   match[2],
			}
			state.process(msg)
		} else {
			logger.Debug("stderr", zap.String("line", line))
		}
	}

	if err := buffer.Err(); err != nil {
		logger.Error("Unable to read stderr", zap.Error(err))
	}
}

type state struct {
	running bool
	cb      FfmpegCallbacks
	logger  *zap.Logger
}

func (s *state) process(msg ComponentMessage) {
	switch msg.Component {
	case "segment":
		matches := openingFilePattern.FindStringSubmatch(msg.Payload)
		if matches != nil {
			path := matches[1]
			if !s.running {
				s.cb.onDownloadStarted()
				s.running = true
			}
			s.cb.onStartWriting(path)
			return
		}

	case "silencedetect":
		startMatches := silenceStartPattern.FindStringSubmatch(msg.Payload)
		if startMatches != nil {
			start, err := parseSeconds(startMatches[1])
			if err == nil {
				s.cb.onSilenceStart(start)
			} else {
				s.logger.Warn("Unable to parse silence start time", zap.Any("msg", msg), zap.Error(err))
			}
			return
		}
		endMatches := silenceEndPattern.FindStringSubmatch(msg.Payload)
		if endMatches != nil {
			end, err := parseSeconds(endMatches[1])
			if err != nil {
				s.logger.Warn("Unable to parse silence end time", zap.Any("msg", msg), zap.Error(err))
				return
			}

			duration, err := parseSeconds((endMatches[2]))
			if err != nil {
				s.logger.Warn("Unable to parse silence duration", zap.Any("msg", msg), zap.Error(err))
				return
			}

			s.cb.onSilenceEnd(end, duration)

			return
		}

	case "out#0/segment":
		// End of output
		return
	}

	s.cb.onUnknownMessage(msg)
}

type ComponentMessage struct {
	Timestamp time.Time
	Component string
	Payload   string
}

type FfmpegCallbacks struct {
	DownloadStarted func()
	StartWriting    func(path string)
	SilenceStart    func(t time.Duration)
	SilenceEnd      func(t time.Duration, duration time.Duration)
	UnknownMessage  func(msg ComponentMessage)
}

func (c *FfmpegCallbacks) onDownloadStarted() {
	if c.DownloadStarted != nil {
		c.DownloadStarted()
	}
}

func (c *FfmpegCallbacks) onStartWriting(path string) {
	if c.StartWriting != nil {
		c.StartWriting(path)
	}
}

func (c *FfmpegCallbacks) onSilenceStart(t time.Duration) {
	if c.SilenceStart != nil {
		c.SilenceStart(t)
	}
}

func (c *FfmpegCallbacks) onSilenceEnd(t time.Duration, duration time.Duration) {
	if c.SilenceEnd != nil {
		c.SilenceEnd(t, duration)
	}
}

func (c *FfmpegCallbacks) onUnknownMessage(msg ComponentMessage) {
	if c.UnknownMessage != nil {
		c.UnknownMessage(msg)
	}
}

func parseSeconds(secs string) (time.Duration, error) {
	s, err := strconv.ParseFloat(secs, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse %q as a duration: %w", secs, err)
	}

	return time.Duration(s*1000*1000*1000) * time.Nanosecond, nil
}
