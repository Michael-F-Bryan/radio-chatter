package graphql

import (
	_ "embed"

	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Schema contains the schema exposed by the GraphQL API.
//
//go:embed schema.graphql
var Schema string

type Resolver struct {
	DB      *gorm.DB
	Logger  *zap.Logger
	Storage radiochatter.BlobStorage
}
