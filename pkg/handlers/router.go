package handlers

import (
	"context"
	"net/http"
	"strconv"

	gql "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	radiochatter "github.com/Michael-F-Bryan/radio-chatter/pkg"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Router(logger *zap.Logger, db *gorm.DB, storage radiochatter.BlobStorage) http.Handler {
	r := mux.NewRouter()

	resolver := &graphql.Resolver{
		DB:      db,
		Logger:  logger.Named("graphql"),
		Storage: storage,
	}
	srv := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{
		Resolvers: resolver,
		Directives: graphql.DirectiveRoot{
			Authenticated: isAuthenticatedDirective,
		},
	}))
	srv.SetRecoverFunc(recoverFunc)
	srv.AddTransport(&transport.Websocket{})
	srv.AroundResponses(logGraphQLErrors)

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

func isAuthenticatedDirective(ctx context.Context, obj interface{}, next gql.Resolver) (res interface{}, err error) {
	logger := GetLogger(ctx)
	logger.Info("Checking auth", zap.Any("obj", obj))
	// TODO: Actually check authentication
	return next(ctx)
}

func recoverFunc(ctx context.Context, err interface{}) (userMessage error) {
	logger := GetLogger(ctx)

	op := gql.GetOperationContext(ctx)

	logger.Error(
		"A GraphQL resolver panicked",
		zap.String("op", op.OperationName),
		zap.String("query", op.RawQuery),
		zap.Any("variables", op.Variables),
		zap.Any("payload", err),
	)

	return gqlerror.Errorf("Internal server error!")
}

func logGraphQLErrors(ctx context.Context, next gql.ResponseHandler) *gql.Response {
	logger := GetLogger(ctx)
	op := gql.GetOperationContext(ctx)

	response := next(ctx)

	if len(response.Errors) > 0 {
		logger.Warn(
			"Resolve error",
			zap.Stringer("path", response.Path),
			zap.String("op", op.OperationName),
			zap.String("query", op.RawQuery),
			zap.Any("variables", op.Variables),
			zap.Error(response.Errors),
		)
	}

	return response
}
