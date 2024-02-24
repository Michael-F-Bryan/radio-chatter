package on_disk_storage

import (
	"net/http"

	"github.com/Michael-F-Bryan/radio-chatter/pkg/middleware"
	"github.com/gorilla/handlers"
	"go.uber.org/zap"
)

func server(logger *zap.Logger, rootDir string) http.Handler {
	return middleware.Apply(
		http.FileServer(http.Dir(rootDir)),
		middleware.Recover(logger.Named("panics")),
		middleware.RequestID,
		middleware.Logging(logger),
		handlers.CORS(),
	)
}
