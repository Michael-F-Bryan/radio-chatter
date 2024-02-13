package handlers

import (
	"net/http"
	"strconv"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Router(logger *zap.Logger, db *gorm.DB) http.Handler {
	r := mux.NewRouter()

	srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{Resolvers: &graphql.Resolver{DB: db}}))

	r.Path("/healthz").Methods(http.MethodHead, http.MethodGet).Handler(Healthz(db, logger))

	graphqlRouter := r.PathPrefix("/graphql").Subrouter()
	graphqlRouter.Methods(http.MethodHead, http.MethodGet).Handler(playground.Handler("GraphQL playground", "/graphql"))
	graphqlRouter.Methods(http.MethodPost).Handler(srv)
	graphqlRouter.Path("schema.graphql").HandlerFunc(graphqlSchema)

	return applyMiddleware(
		r,
		handlers.RecoveryHandler(handlers.RecoveryLogger(&recoveryLogger{logger: logger.Named("panics")})),
		handlers.CompressHandler,
		requestIDMiddleware,
		loggingMiddleware(logger),
		handlers.CORS(
			handlers.AllowedMethods([]string{http.MethodHead, http.MethodGet, http.MethodPost}),
			// Note: It's fine to allow all origins because that lets users access
			// the backend from their own apps
			handlers.AllowedOrigins([]string{"*"}),
		),
	)
}

func graphqlSchema(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")
	w.Header().Add("Content-Disposition", "attachment; filename='schema.graphql'")
	w.Header().Add("Content-Length", strconv.Itoa(len(graphql.Schema)))

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(graphql.Schema))
}

// applyMiddleware will wrap a handler in a series of middleware functions,
// taking care to make sure the middleware is executed in the order they are
// provided.
func applyMiddleware(handler http.Handler, middleware ...mux.MiddlewareFunc) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		m := middleware[i]
		handler = m(handler)
	}
	return handler
}

type recoveryLogger struct {
	logger *zap.Logger
}

func (r *recoveryLogger) Println(values ...any) {
	var payload any

	if len(values) == 1 {
		payload = values[0]
	} else {
		payload = values
	}

	r.logger.Error("A panic occurred", zap.Any("payload", payload))
}
