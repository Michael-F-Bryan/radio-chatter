package main

import (
	"context"
	"errors"
	"time"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
	glog "gorm.io/gorm/logger"
)

func registerDatabaseFlags(flags *pflag.FlagSet) {
	flags.String("db", "radio-chatter.sqlite3", "The database to save to")
	_ = viper.BindPFlag("db.source", flags.Lookup("db"))
	_ = viper.BindEnv("db.source", "DB_SOURCE")

	flags.String("db-driver", "sqlite3", "Which database type to use")
	_ = viper.BindPFlag("db.driver", flags.Lookup("db-driver"))
	_ = viper.BindEnv("db.driver", "DB_DRIVER")

	flags.Bool("trace", false, "Trace all SQL queries")
	_ = viper.BindPFlag("db.trace", flags.Lookup("trace"))
	_ = viper.BindEnv("db.trace", "DB_TRACE")
}

func setupDatabase(ctx context.Context, logger *zap.Logger, cfg Config) *gorm.DB {
	if logger.Name() != "db" {
		logger = logger.Named("db")
	}

	db, err := radiochatter.OpenDatabase(ctx, logger, cfg.Database.Driver, cfg.Database.Source)
	if err != nil {
		logger.Fatal("Unable to initialize the database", zap.Error(err))
	}

	l := &zapLogger{
		inner:    logger,
		sugar:    logger.Sugar(),
		traceSQL: cfg.Database.Trace,
		logLevel: glog.Silent,
	}
	if cfg.Output.DevMode {
		l.logLevel = glog.Info
	}
	db.Logger = l

	return db
}

type zapLogger struct {
	inner    *zap.Logger
	sugar    *zap.SugaredLogger
	traceSQL bool
	logLevel glog.LogLevel
}

func (l *zapLogger) LogMode(level glog.LogLevel) glog.Interface {
	logger := *l
	logger.logLevel = level
	return &logger
}

func (l *zapLogger) Info(ctx context.Context, fmt string, args ...interface{}) {
	if l.logLevel >= glog.Info {
		args = append([]any{fmt}, args...)
		l.sugar.Info(args...)
	}
}

func (l *zapLogger) Warn(ctx context.Context, fmt string, args ...interface{}) {
	if l.logLevel >= glog.Warn {
		args = append([]any{fmt}, args...)
		l.sugar.Info(args...)
	}
}

func (l *zapLogger) Error(ctx context.Context, fmt string, args ...interface{}) {
	if l.logLevel >= glog.Warn {
		args = append([]any{fmt}, args...)
		l.sugar.Info(args...)
	}
}

func (l *zapLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		sql, rowsAffected := fc()
		l.inner.Error(
			"SQL Error",
			zap.Time("begin", begin),
			zap.String("sql", sql),
			zap.Int64("rows-affected", rowsAffected),
			zap.Error(err),
		)
	} else if l.traceSQL {
		sql, rowsAffected := fc()
		l.inner.Debug(
			"SQL trace",
			zap.Time("begin", begin),
			zap.Duration("duration", time.Since(begin)),
			zap.String("sql", sql),
			zap.Int64("rows-affected", rowsAffected),
		)
	}
}
