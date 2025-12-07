package ports

import (
	"context"
	"io"
)

type ImageStore interface {
	UploadImage(ctx context.Context, imageData io.Reader, name string) (url string, err error)
	DeleteImage(ctx context.Context, imageUrl string) error
  GetImageDownloadURL(ctx context.Context, imageUrl string) (url string, err error)
}
