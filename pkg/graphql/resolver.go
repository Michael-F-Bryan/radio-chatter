package graphql

import (
	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Resolver struct {
	DB      *gorm.DB
	Logger  *zap.Logger
	Storage radiochatter.BlobStorage
}
