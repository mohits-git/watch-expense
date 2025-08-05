package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/mohits-git/watch-expense/internal/app"
)

type Handler struct {
	app *app.App
}

func NewHandler(app *app.App) *Handler {
	return &Handler{
		app: app,
	}
}

func (h *Handler) ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("docs", "openapi.yaml"))
}

func (h *Handler) SwaggerUIHandler() http.Handler {
	return http.StripPrefix("/docs/", http.FileServer(http.Dir("docs/swagger")))
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
