package handlers

import (
	"net/http"
	"strconv"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Router(logger *zap.Logger, db *gorm.DB) http.Handler {
	r := mux.NewRouter()

	resolver := &graphql.Resolver{
		DB:     db,
		Logger: logger.Named("graphql"),
	}
	srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{Resolvers: resolver}))
	srv.AddTransport(&transport.Websocket{})

	r.Path("/healthz").Methods(http.MethodHead, http.MethodGet).Handler(Healthz(db, logger))
	r.Path("/graphql").Schemes("http", "https", "ws", "wss").Handler(srv)
	r.Path("/graphql/playground").Handler(playground.Handler("GraphQL playground", "/graphql"))
	r.Path("/graphql/schema.graphql").Methods(http.MethodHead, http.MethodGet).HandlerFunc(graphqlSchema)

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
