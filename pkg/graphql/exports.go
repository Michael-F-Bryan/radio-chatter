package graphql

import (
	_ "embed"

	gql "github.com/99designs/gqlgen/graphql"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql/generated"
)

type Config = generated.Config
type DirectiveRoot = generated.DirectiveRoot

func NewExecutableSchema(cfg Config) gql.ExecutableSchema { return generated.NewExecutableSchema(cfg) }

// Schema contains the schema exposed by the GraphQL API.
//
//go:embed schema.graphql
var Schema string
