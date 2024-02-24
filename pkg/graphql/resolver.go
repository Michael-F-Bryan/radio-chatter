package graphql

import (
	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"gorm.io/gorm"
)

type Resolver struct {
	DB      *gorm.DB
	Storage blob.Storage
}
