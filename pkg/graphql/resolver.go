package graphql

import (
	_ "embed"

	"gorm.io/gorm"
)

// Schema contains the schema exposed by the GraphQL API.
//
//go:embed schema.graphql
var Schema string

type Resolver struct {
	DB *gorm.DB
}
