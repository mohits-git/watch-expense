package ports

import (
	"context"
	"io"
)

type ImageUploadService interface {
	UploadImage(ctx context.Context, imageData io.Reader) (url string, err error)
	DeleteImage(ctx context.Context, imageUrl string) error
}
