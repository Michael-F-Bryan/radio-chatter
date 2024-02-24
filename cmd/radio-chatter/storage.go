package main

import (
	"os"
	"path"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/on_disk_storage"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func registerStorageFlags(flags *pflag.FlagSet) {
	flags.StringP("blob", "b", "", "Blob storage (user's cache dir by default)")
	_ = viper.BindPFlag("blob", flags.Lookup("blob"))
	_ = viper.BindEnv("storage.blob", "BLOB_URL")
}

func setupStorage(logger *zap.Logger, cfg StorageConfig) blob.Storage {
	if logger.Name() != "storage" {
		logger = logger.Named("storage")
	}

	baseDir := cfg.Blob

	if baseDir == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			logger.Fatal("Unable to get the user's cache directory", zap.Error(err))
		}
		baseDir = path.Join(cacheDir, "radio-chatter", "blob-storage")
	}

	storage, err := on_disk_storage.New(logger, baseDir)
	if err != nil {
		logger.Fatal("Unable to set up the on-disk storage", zap.Error(err))
	}

	return storage
}
