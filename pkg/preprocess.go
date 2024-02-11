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
const clipLength = 60 * time.Second

// messagePattern will match a string like "[silencedetect @ 0x600000c583c0] ..."
var messagePattern = regexp.MustCompile(`^\[(\S+) @ [\d\w]+\]\s*(.*)$`)

// openingFilePattern matches a string like "Opening '/path/to/file.mp3' for writing"
var openingFilePattern = regexp.MustCompile(`^Opening '([^']+)' for writing$`)

var silenceStartPattern = regexp.MustCompile(`silence_start: (\d+(?:\.\d+)?)`)
var silenceEndPattern = regexp.MustCompile(`silence_end: (\d+(?:\.\d+)?) \| silence_duration: (\d+(?:\.\d+)?)`)

// Preprocess will take some ffmpeg input and split it into 60-second chunks,
// saved in the provided directory.
//
// The input may be a URL or a filename.
//
// Ffmpeg will try to gracefully shut down (flushing buffers, terminating mp3,
// etc.) when the context is cancelled. However, if the graceful shutdown
// doesn't complete within a reasonable amount of time it will be forcefully
// killed.
func Preprocess(ctx context.Context, logger *zap.Logger, input string, outputDir string, cb PreprocessingCallbacks) error {
	args := []string{
		"-i", input,
		// Use a filter to detect silence and print its timestamps
		"-af", "silencedetect=noise=-30dB:d=1",
		// Split into 60-second chunks
		"-f", "segment", "-segment_time", strconv.Itoa(int(clipLength) / int(time.Second)),
		// Clean up stderr so it's easier to parse
		"-hide_banner", "-nostdin", "-nostats",
		// the output path
		path.Join(outputDir, "chunk_%d.mp3"),
	}

	cmd := exec.CommandContext(ctx, defaultCommand, args...)
	cmd.Stdout = os.Stdout

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	parsingFinished := make(chan error, 1)
	go func() {
		defer close(parsingFinished)
		parsingFinished <- parseStderr(logger, stderr, cb)
	}()

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
		"ffmpeg started",
		zap.Stringer("cmd", cmd),
		zap.Int("pid", cmd.Process.Pid),
	)

	err = cmd.Wait()

	exitCode := cmd.ProcessState.ExitCode()
	logger.Debug("ffmpeg exited", zap.Int("exit-code", exitCode))

	// As an edge case, ffmpeg will exit with a 255 exit code when you send it a
	// SIGINT (e.g. because we told it to shut down). This is an expected
	// condition, so don't treat it like an error.
	if err != nil && exitCode == 255 {
		err = nil
	}

	// Note: we want to make sure parsing has finished and no more callbacks are
	// triggered before returning from this function. That way we don't end
	// up with any dangling goroutines and everything is deterministic.
	parsingError := <-parsingFinished

	// There are two possible sources of errors, 1) parsing could have ran into
	// problems (e.g. invalid UTF-8), and 2) the process could have exited with
	// a non-zero exit code. We want to prefer telling users about the latter
	// because it's probably more actionable.
	if err != nil {
		return fmt.Errorf("ffmpeg exited unsucessfully: %w", err)
	} else {
		return parsingError
	}
}

// parseStderr reads the output from ffmpeg and triggers callbacks to notify
// the caller when certain events occur.
func parseStderr(logger *zap.Logger, stderr io.Reader, cb PreprocessingCallbacks) error {
	state := state{
		running: false,
		cb:      cb,
		logger:  logger,
	}
	buffer := bufio.NewScanner(stderr)

	defer cb.onFinished()

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
			cb.onUninterpretedStderr(line)
		}
	}

	return buffer.Err()
}

type state struct {
	running bool
	cb      PreprocessingCallbacks
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
		// End of input
		return
	}

	s.cb.onUnknownMessage(msg)
}

type ComponentMessage struct {
	Timestamp time.Time
	Component string
	Payload   string
}

// PreprocessingCallbacks are events used to notify the caller when certain
// preprocessing events occur.
//
// Callbacks must be goroutine-safe.
//
// Once a call to Preprocess() has completed, no further callbacks will be
// called.
type PreprocessingCallbacks struct {
	// Ffmpeg has started downloading the input file.
	DownloadStarted func()
	// The preprocessor has started writing to a new chunk.
	StartWriting func(path string)
	// The start of a silent period has been detected.
	//
	// The duration is relative to the start of the input.
	SilenceStart func(t time.Duration)
	// The end of a silent period has been detected.
	//
	// The durations are relative to the start of the input.
	SilenceEnd func(t time.Duration, duration time.Duration)
	// An unknown message type was encountered.
	UnknownMessage func(msg ComponentMessage)
	// Received a line on stderr that wasn't part of a message.
	UninterpretedStderr func(string)
	// Preprocessing has finished.
	Finished func()
}

func (c *PreprocessingCallbacks) onDownloadStarted() {
	if c.DownloadStarted != nil {
		c.DownloadStarted()
	}
}

func (c *PreprocessingCallbacks) onStartWriting(path string) {
	if c.StartWriting != nil {
		c.StartWriting(path)
	}
}

func (c *PreprocessingCallbacks) onSilenceStart(t time.Duration) {
	if c.SilenceStart != nil {
		c.SilenceStart(t)
	}
}

func (c *PreprocessingCallbacks) onSilenceEnd(t time.Duration, duration time.Duration) {
	if c.SilenceEnd != nil {
		c.SilenceEnd(t, duration)
	}
}

func (c *PreprocessingCallbacks) onUnknownMessage(msg ComponentMessage) {
	if c.UnknownMessage != nil {
		c.UnknownMessage(msg)
	}
}

func (c *PreprocessingCallbacks) onUninterpretedStderr(stderr string) {
	if c.UninterpretedStderr != nil {
		c.UninterpretedStderr(stderr)
	} else {
		zap.L().Debug("stderr", zap.String("line", stderr))
	}
}

func (c *PreprocessingCallbacks) onFinished() {
	if c.Finished != nil {
		c.Finished()
	}
}

func parseSeconds(secs string) (time.Duration, error) {
	s, err := strconv.ParseFloat(secs, 64)
	if err != nil {
		return 0, fmt.Errorf("unable to parse %q as a duration: %w", secs, err)
	}

	return time.Duration(s*1000*1000*1000) * time.Nanosecond, nil
}
