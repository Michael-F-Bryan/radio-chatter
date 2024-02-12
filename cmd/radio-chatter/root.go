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

var DevMode bool
var Output = TextFormat

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "radio-chatter",
		Short:             "A brief description of your application",
		Long:              "",
		PersistentPreRun:  beforeAll,
		PersistentPostRun: afterAll,
	}

	cmd.AddCommand(downloadCmd())

	flags := cmd.PersistentFlags()
	flags.BoolVarP(&DevMode, "dev", "d", false, "Run the application in dev mode")
	_ = viper.BindPFlag("dev", flags.Lookup("dev"))

	return cmd
}

func beforeAll(cmd *cobra.Command, args []string) {
	initializeLogger()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	cmd.SetContext(ctx)

	post := cmd.PersistentPostRun
	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		cancel()
		if post != nil {
			post(cmd, args)
		}
	}

	go func() {
		<-ctx.Done()
		zap.L().Debug("Beginning graceful shutdown (press ctrl-C to exit forcefully)")
		cancel()
	}()

	viper.SetEnvPrefix("rc")
	viper.AutomaticEnv()

	if userConfigDir, err := os.UserConfigDir(); err == nil {
		configDir := path.Join(userConfigDir, radiochatter.PackageIdentifier)
		viper.AddConfigPath(configDir)
	}
}

func initializeLogger() {
	var cfg zap.Config

	if DevMode {
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
