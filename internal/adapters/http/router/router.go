package router

import (
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/http/handlers"
)

func NewHTTPRouter(
	commonHandler *handlers.CommonHandler,
) http.Handler {
	mux := http.NewServeMux()

	// docs
	mux.HandleFunc("/docs/openapi.yaml", commonHandler.ServeOpenAPISpec)
	mux.Handle("/docs/", commonHandler.SwaggerUIHandler())

	// health check
	mux.HandleFunc("/health", commonHandler.HealthCheck)

	// Add more routes...

	return mux
}
