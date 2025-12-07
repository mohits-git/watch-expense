package handlers

import (
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

type ImageUploadHandler struct {
	imageUploadService services.ImageService
}

func NewImageUploadHandler(service services.ImageService) *ImageUploadHandler {
	return &ImageUploadHandler{
		imageUploadService: service,
	}
}

func (h *ImageUploadHandler) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Failed to parse multipart form")
	}

	file, header, err := r.FormFile("file")
	filename := header.Filename
	defer file.Close()

	url, err := h.imageUploadService.UploadUserImage(r.Context(), file, filename)
	if err != nil {
		if apperr.IsTooLargeError(err) {
			writeError(w, 413, "File too large")
			return
		}
		writeError(w, http.StatusInternalServerError, "Something went wrong while uploading your image.")
		return
	}
	response := dtos.ImageUploadResponse{ImageURL: url}
	writeResponse(w, http.StatusCreated, "Successfully uploaded image.", response)
}

func (h *ImageUploadHandler) HandleDeleteImage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	deleteReq, err := decodeRequest[dtos.DeleteImageRequest](r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	url := deleteReq.ImageURL
	err = h.imageUploadService.DeleteUserImage(r.Context(), url)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not delete the image")
		return
	}
	writeResponse(w, http.StatusOK, "Successfully deleted image.", struct{}{})
}
