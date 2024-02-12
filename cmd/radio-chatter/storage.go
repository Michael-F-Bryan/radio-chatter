package main

import (
	"os"
	"path"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var StorageDir string

func registerStorageFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&StorageDir, "blob", "b", "", "Blob storage (user's cache dir by default)")
	_ = viper.BindPFlag("blob", flags.Lookup("blob"))
}

func setupStorage(logger *zap.Logger) radiochatter.BlobStorage {
	baseDir := StorageDir

	if baseDir == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			logger.Error("Unable to get the user's cache directory", zap.Error(err))
		}
		baseDir = path.Join(cacheDir, "radio-chatter", "blob-storage")
	}

	return radiochatter.NewOnDiskStorage(logger, baseDir)
}
