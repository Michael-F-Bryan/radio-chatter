package handler

import (
	"net/http"
	"strconv"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Router(logger *zap.Logger, db *gorm.DB) http.Handler {
	r := mux.NewRouter()

	r.Path("/graphql").Methods(http.MethodGet).Handler(playground.Handler("GraphQL playground", "/graphql"))

	srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{Resolvers: &graphql.Resolver{DB: db}}))
	r.Path("/graphql").Methods(http.MethodPost).Handler(srv)

	r.Path("/graphql/schema.graphql").Methods(http.MethodGet).HandlerFunc(graphqlSchema)

	return r
}

func graphqlSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	w.Header().Add("Content-Disposition", "attachment; filename='schema.graphql'")
	w.Header().Add("Content-Length", strconv.Itoa(len(graphql.Schema)))

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(graphql.Schema))
}
