package main

import (
	"context"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var DbDriver string
var ConnectionString string

func registerDatabaseFlags(flags *pflag.FlagSet) {
	flags.StringVar(&ConnectionString, "db", "radio-chatter.sqlite3", "The database to save to")
	viper.BindPFlag("db", flags.Lookup("db"))
	flags.StringVar(&DbDriver, "db-driver", "sqlite3", "Which database type to use")
	viper.BindPFlag("db-driver", flags.Lookup("db-driver"))
}

func setupDatabase(ctx context.Context, logger *zap.Logger) (*gorm.DB, error) {
	return radiochatter.OpenDatabase(ctx, logger, DbDriver, ConnectionString)
}
