package graphql

import (
	_ "embed"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Schema contains the schema exposed by the GraphQL API.
//
//go:embed schema.graphql
var Schema string

type Resolver struct {
	DB     *gorm.DB
	Logger *zap.Logger
}
