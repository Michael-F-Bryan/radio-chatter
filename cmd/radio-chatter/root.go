package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "radio-chatter",
		Short:             "A brief description of your application",
		Long:              "",
		PersistentPreRun:  beforeAll,
		PersistentPostRun: afterAll,
	}

	cmd.AddCommand(downloadCmd(), streamCmd(), serveCmd(), configCmd())

	flags := cmd.PersistentFlags()
	flags.BoolP("dev", "d", false, "Run the application in dev mode")
	_ = viper.BindPFlag("dev", flags.Lookup("dev"))

	registerFormatFlags(cmd.PersistentFlags())

	return cmd
}

func beforeAll(cmd *cobra.Command, args []string) {
	initializeLogger()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	viper.SetConfigName("radio-chatter")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if userConfigDir, err := os.UserConfigDir(); err == nil {
		configDir := path.Join(userConfigDir, radiochatter.PackageIdentifier)
		viper.AddConfigPath(configDir)
	}

	_ = viper.ReadInConfig()

	var cfg Config

	if err := viper.Unmarshal(&cfg); err != nil {
		zap.L().Fatal("Unable to load the config", zap.Error(err))
	}
	ctx = context.WithValue(ctx, configKey{}, cfg)
	zap.L().Debug("Loaded config", zap.Any("settings", cfg))

	post := cmd.PersistentPostRun
	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		cancel()
		if post != nil {
			post(cmd, args)
		}
	}

	go func() {
		<-ctx.Done()
		// Cancel the signal handler so pressing ctrl-C again will kill the
		// program.
		cancel()
	}()

	cmd.SetContext(ctx)
}

func initializeLogger() {
	var cfg zap.Config

	if viper.GetBool("dev") {
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

func afterAll(cmd *cobra.Command, args []string) {
	_ = zap.L().Sync()
}
