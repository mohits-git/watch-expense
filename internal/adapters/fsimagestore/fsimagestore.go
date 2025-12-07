package fsimagestore

import (
	"context"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/ports"
)

type FSImageStore struct {
	baseUrl             string
	fileUploadDirectory string
}

func NewFSImageStore(baseUrl, fileUploadDirectory string) ports.ImageStore {
	// create the directory if it doesn't exist
	if err := os.MkdirAll(fileUploadDirectory, os.ModePerm); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	return &FSImageStore{
		baseUrl:             baseUrl,
		fileUploadDirectory: fileUploadDirectory,
	}
}

func (s *FSImageStore) UploadImage(ctx context.Context, imageData io.Reader, name string) (string, error) {
	id := uuid.New().String() + "_" + name
	filePath := filepath.Join(s.fileUploadDirectory, id)
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, imageData)
	if err != nil {
		return "", err
	}

	imageUrl, err := url.JoinPath(s.baseUrl, filePath)
	if err != nil {
		return "", err
	}
	return imageUrl, nil
}

func (s *FSImageStore) DeleteImage(ctx context.Context, imageUrl string) error {
	parsedUrl, err := url.Parse(imageUrl)
	if err != nil {
		return err
	}

	filePath := strings.TrimPrefix(parsedUrl.Path, "/")
	err = os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}

func (s *FSImageStore) GetImageDownloadURL(ctx context.Context, imageUrl string) (string, error) {
  // In a filesystem-based store, the upload URL is the same as the download URL
  return imageUrl, nil
}
