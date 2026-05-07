package routes

import (
	"net/http"

	"github.com/MorseWayne/gogopher-arch/src/internal/config"
	"github.com/MorseWayne/gogopher-arch/src/services/gateway/internal/handlers"
	"github.com/MorseWayne/gogopher-arch/src/services/gateway/internal/middleware"
)

func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/execute", handlers.NewExecuteHandler(cfg))
	mux.Handle("/health", handlers.NewHealthHandler())

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	return h
}