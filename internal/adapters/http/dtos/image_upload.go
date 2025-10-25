package dtos

type ImageUploadResponse struct {
	ImageURL string `json:"image_url"`
}

type DeleteImageRequest struct {
	ImageURL string `json:"image_url"`
}
