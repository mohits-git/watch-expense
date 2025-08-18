package router

import (
	"net/http"

	"github.com/mohits-git/watch-expense/internal/app"
	"github.com/mohits-git/watch-expense/internal/handlers"
)

type Router struct {
	app *app.App
}

func NewRouter(app *app.App) *Router {
	return &Router{
		app: app,
	}
}

func (r *Router) Route() *http.ServeMux {
	mux := http.NewServeMux()
	// Initialize the handler with the app instance
	h := handlers.NewHandler(r.app)

	// docs
	mux.HandleFunc("/docs/openapi.yaml", h.ServeOpenAPISpec)
	mux.Handle("/docs/", h.SwaggerUIHandler())

	// health check
	mux.HandleFunc("/health", h.HealthCheck)

	// Add more routes...

	return mux
}
