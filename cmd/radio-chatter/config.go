package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type configKey struct{}

func GetConfig(ctx context.Context) Config {
	cfg, ok := ctx.Value(configKey{}).(Config)

	if !ok {
		panic("A config hasn't been attached to this context")
	}

	return cfg

}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Print out the current config",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := GetConfig(cmd.Context())
			if err := cfg.Format().Print(os.Stdout, cfg); err != nil {
				zap.L().Fatal("Unable to print the config", zap.Error(err))
			}
		},
	}

	return cmd
}

type Config struct {
	Serve    ServeConfig    `mapstructure:"serve" json:"serve"`
	Storage  StorageConfig  `mapstructure:"storage" json:"storage"`
	Database DatabaseConfig `mapstructure:"db" json:"db"`
	Output   OutputConfig   `mapstructure:"out" json:"out"`
}

func (c Config) Format() formatter {
	name := c.Output.Format
	if name == "" {
		name = "text"
	}
	f, err := getFormatter(name)
	if err != nil {
		zap.L().Warn("Unable to get the formatter", zap.Error(err))
	}
	return f
}

type ServeConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port uint16 `mapstructure:"port" json:"port"`
}

func (s ServeConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type StorageConfig struct {
	Blob string `mapstructure:"blob" json:"blob"`
}

type DatabaseConfig struct {
	Source string `mapstructure:"source" json:"source"`
	Driver string `mapstructure:"driver" json:"driver"`
	Trace  bool   `mapstructure:"trace" json:"trace"`
}

type OutputConfig struct {
	DevMode bool   `mapstructure:"dev" json:"dev"`
	Format  string `mapstructure:"format" json:"format"`
}
