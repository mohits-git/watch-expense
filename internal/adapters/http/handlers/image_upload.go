package handlers

import (
	"log"
	"net/http"

	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type ImageUploadHandler struct {
	imageUploadService ports.ImageUploadService
}

func NewImageUploadHandler(service ports.ImageUploadService) *ImageUploadHandler {
	return &ImageUploadHandler{
		imageUploadService: service,
	}
}

func (h *ImageUploadHandler) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	url, err := h.imageUploadService.UploadImage(r.Context(), r.Body)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	url := deleteReq.ImageURL
	// TODO: check if the user is the owner of the uploaded image
	err = h.imageUploadService.DeleteImage(r.Context(), url)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusInternalServerError, "Could not delete the image")
		return
	}
	writeResponse(w, http.StatusOK, "Successfully deleted image.", struct{}{})
}

