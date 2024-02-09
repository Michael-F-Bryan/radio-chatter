package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
)

func main() {
	args := parseArgs()
	args.initializeLogger()
	logger := zap.L()
	defer func() { _ = logger.Sync() }()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	group, ctx := errgroup.WithContext(ctx)

	ch := make(chan radiochatter.ComponentMessage)
	go func() {
		for msg := range ch {
			logger.Info("message", zap.Any("msg", msg))
		}
	}()

	dir, err := os.MkdirTemp("", "radio-chatter-tmp")
	if err != nil {
		logger.Fatal("Unable to create a temporary directory", zap.Error(err))
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			logger.Error(
				"Unable to remove the temporary directory",
				zap.String("temp", dir),
				zap.Error(err),
			)
		}
	}()

	group.Go(func() error {
		cb := callbacks{logger: logger}
		return radiochatter.RunFfmpeg(ctx, logger, args.url, dir, cb.Callbacks())
	})

	logger.Info("Starting", zap.String("temp", dir))
	defer logger.Info("Shutting down")

	if err := group.Wait(); err != nil {
		logger.Fatal("Failed", zap.Error(err))
	}
}

type args struct {
	devMode bool
	url     string
}

func parseArgs() args {
	var args args

	flag.BoolVar(&args.devMode, "dev", false, "Enable dev mode")
	flag.StringVar(&args.url, "url", radiochatter.FeedUrl, "The feed to download")
	flag.Parse()

	return args
}

func (a args) initializeLogger() {
	var cfg zap.Config

	if a.devMode {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		log.Fatalf("Unable to initialize the logger: %e", err)
	}

	zap.ReplaceGlobals(logger)
}

type callbacks struct {
	logger *zap.Logger
}

func (c *callbacks) DownloadStarted() {
	c.logger.Info("Download started")
}

func (c *callbacks) StartWriting(path string) {
	c.logger.Info("Started writing", zap.String("path", path))
}

func (c *callbacks) UnknownMessage(msg radiochatter.ComponentMessage) {
	c.logger.Debug("Message", zap.Any("msg", msg))
}

func (c *callbacks) SilenceStart(t time.Duration) {
	c.logger.Info("Silence started", zap.Duration("start", t))
}

func (c *callbacks) SilenceEnd(t time.Duration, duration time.Duration) {
	c.logger.Info("Silence ended", zap.Duration("end", t), zap.Duration("duration", duration))
}

func (c *callbacks) Callbacks() radiochatter.FfmpegCallbacks {
	return radiochatter.FfmpegCallbacks{
		DownloadStarted: c.DownloadStarted,
		StartWriting:    c.StartWriting,
		SilenceStart:    c.SilenceStart,
		SilenceEnd:      c.SilenceEnd,
		UnknownMessage:  c.UnknownMessage,
	}
}
