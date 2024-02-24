package handlers

import (
	"context"
	"net/http"
	"net/http/pprof"
	"strconv"

	gql "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/blob"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/graphql"
	"github.com/Michael-F-Bryan/radio-chatter/pkg/middleware"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Router(logger *zap.Logger, db *gorm.DB, storage blob.Storage, devMode bool) http.Handler {
	r := mux.NewRouter()

	resolver := &graphql.Resolver{
		DB:      db,
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

	r.Path("/healthz").Methods(http.MethodHead, http.MethodGet).Handler(Healthz(db))
	r.Path("/graphql").Schemes("http", "https", "ws", "wss").Methods(http.MethodHead, http.MethodGet, http.MethodPost, http.MethodOptions).Handler(srv)
	r.Path("/graphql/playground").Handler(playground.Handler("GraphQL playground", "/graphql"))
	r.Path("/graphql/schema.graphql").Methods(http.MethodHead, http.MethodGet).HandlerFunc(graphqlSchema)

	if devMode {
		logger.Info("Registering debug endpoints")
		sub := r.PathPrefix("/debug/pprof").Subrouter()
		sub.HandleFunc("/cmdline", pprof.Cmdline)
		sub.HandleFunc("/profile", pprof.Profile)
		sub.HandleFunc("/symbol", pprof.Symbol)
		sub.HandleFunc("/trace", pprof.Trace)
		sub.NotFoundHandler = http.HandlerFunc(pprof.Index)
	}

	return middleware.Apply(
		r,
		middleware.Recover(logger.Named("panics")),
		handlers.CompressHandler,
		middleware.RequestID,
		middleware.Logging(logger),
		handlers.CORS(
			// Note: It's fine to allow all origins because that lets users access
			// the backend from their own apps
			handlers.AllowedOrigins([]string{"*"}),
			// Note: Chrome likes to send some extra headers that need to be
			// explicitly allowed
			handlers.AllowedHeaders([]string{"Content-Type", "Origin"}),
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

func isAuthenticatedDirective(ctx context.Context, obj interface{}, next gql.Resolver) (res interface{}, err error) {
	logger := middleware.GetLogger(ctx)
	logger.Info("Checking auth", zap.Any("obj", obj))
	// TODO: Actually check authentication
	return next(ctx)
}

func recoverFunc(ctx context.Context, err interface{}) (userMessage error) {
	logger := middleware.GetLogger(ctx)

	op := gql.GetOperationContext(ctx)

	logger.Error(
		"A GraphQL resolver panicked",
		zap.String("op", op.OperationName),
		zap.String("query", op.RawQuery),
		zap.Any("variables", op.Variables),
		zap.Any("payload", err),
	)

	return gqlerror.Errorf("Resolver panicked: %s", err)
}

func logGraphQLErrors(ctx context.Context, next gql.ResponseHandler) *gql.Response {
	logger := middleware.GetLogger(ctx)
	op := gql.GetOperationContext(ctx)

	response := next(ctx)

	if response != nil && len(response.Errors) > 0 {
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
