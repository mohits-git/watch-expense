package handlers

import (
	"net/http"
	"path/filepath"
)

type CommonHandler struct {
}

func NewCommonHandler() *CommonHandler {
	return &CommonHandler{}
}

func (h *CommonHandler) ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("docs", "openapi.yaml"))
}

func (h *CommonHandler) SwaggerUIHandler() http.Handler {
	return http.StripPrefix("/docs/", http.FileServer(http.Dir("docs/swagger")))
}

func (h *CommonHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
