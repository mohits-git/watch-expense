package mysql

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type ImageMetadataRepository struct {
  db *sql.DB
}

func NewImageMetadataRepository(db *sql.DB) ports.ImageMetadataRepository {
  return &ImageMetadataRepository{db: db}
}

func (imr *ImageMetadataRepository) SaveImageUserMetadata(ctx context.Context, imageURL string, userID string) error {
  id := uuid.New().String()
  query := `INSERT INTO image_metadata (id, image_url, user_id) VALUES (?, ?, ?)`
  _, err := imr.db.ExecContext(ctx, query, id, imageURL, userID)
  if err != nil {
    return err
  }
  return nil
}

func (imr *ImageMetadataRepository) GetImageUserMetadata(ctx context.Context, imageURL string) (userID string, err error) {
  query := `SELECT user_id FROM image_metadata WHERE image_url = ?`
  row := imr.db.QueryRowContext(ctx, query, imageURL)
  err = row.Scan(&userID)
  if err != nil {
    return "", err
  }
  return userID, nil
}

func (imr *ImageMetadataRepository) DeleteImageUserMetadata(ctx context.Context, imageURL string) error {
  query := `DELETE FROM image_metadata WHERE image_url = ?`
  _, err := imr.db.ExecContext(ctx, query, imageURL)
  if err != nil {
    return err
  }
  return nil
}
