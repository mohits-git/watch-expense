package imageupload

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

type FSImageUpload struct {
	baseUrl             string
	fileUploadDirectory string
}

func NewFSImageUpload(baseUrl, fileUploadDirectory string) ports.ImageUploadService {
	return &FSImageUpload{
		baseUrl:             baseUrl,
		fileUploadDirectory: fileUploadDirectory,
	}
}

func (s *FSImageUpload) UploadImage(ctx context.Context, imageData io.Reader) (string, error) {
	id := uuid.New().String()
	filePath := filepath.Join(s.fileUploadDirectory, id)
	log.Println(filePath)
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

func (s *FSImageUpload) DeleteImage(ctx context.Context, imageUrl string) error {
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
