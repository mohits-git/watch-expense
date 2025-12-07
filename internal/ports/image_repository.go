package ports

import (
  "context"
)

type ImageMetadataRepository interface {
  SaveImageUserMetadata(ctx context.Context, imageURL string, userID string) error
  GetImageUserMetadata(ctx context.Context, imageURL string) (userID string, err error)
  DeleteImageUserMetadata(ctx context.Context, imageURL string) error
}
