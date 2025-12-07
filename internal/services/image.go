package services

import (
	"context"
	"io"

	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
)

type ImageService interface {
	UploadUserImage(ctx context.Context, imageData io.Reader, imageName string) (imageURL string, err error)
	DeleteUserImage(ctx context.Context, imageURL string) error
	GetUserImageDownloadURL(ctx context.Context, imageURL string) (downloadURL string, err error)
}

type imageService struct {
	imageStore        ports.ImageStore
	imageMetadataRepo ports.ImageMetadataRepository
}

func NewImageService(imageStore ports.ImageStore, imageMetadataRepo ports.ImageMetadataRepository) ImageService {
	return &imageService{
		imageStore:        imageStore,
		imageMetadataRepo: imageMetadataRepo,
	}
}

func (imageService *imageService) UploadUserImage(ctx context.Context, imageData io.Reader, imageName string) (imageURL string, err error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok || claims == nil || claims.UserID == "" {
		return "", apperr.NewAppError(apperr.ErrUnauthorized, "user not authenticated", nil)
	}
	userID := claims.UserID

	url, err := imageService.imageStore.UploadImage(ctx, imageData, imageName)
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "failed to upload image", err)
	}

	err = imageService.imageMetadataRepo.SaveImageUserMetadata(ctx, url, userID)
	if err != nil {
		_ = imageService.imageStore.DeleteImage(ctx, url)
		return "", apperr.NewAppError(apperr.ErrInternal, "failed to save image metadata", err)
	}
	return url, nil
}

func (imageService *imageService) DeleteUserImage(ctx context.Context, imageURL string) error {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok || claims == nil || claims.UserID == "" {
		return apperr.NewAppError(apperr.ErrUnauthorized, "user not authenticated", nil)
	}
	userID := claims.UserID

	ownerID, err := imageService.imageMetadataRepo.GetImageUserMetadata(ctx, imageURL)
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "failed to get image metadata", err)
	}
	if ownerID != userID {
		return apperr.NewAppError(apperr.ErrForbidden, "user not authorized to delete this image", nil)
	}

	err = imageService.imageMetadataRepo.DeleteImageUserMetadata(ctx, imageURL)
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "failed to delete image metadata", err)
	}

	err = imageService.imageStore.DeleteImage(ctx, imageURL)
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "failed to delete image", err)
	}
	return nil
}

func (imageService *imageService) GetUserImageDownloadURL(ctx context.Context, imageURL string) (downloadURL string, err error) {
	claims, ok := authctx.UserClaimsFromCtx(ctx)
	if !ok || claims == nil || claims.UserID == "" {
		return "", apperr.NewAppError(apperr.ErrUnauthorized, "user not authenticated", nil)
	}
	userID := claims.UserID

	ownerID, err := imageService.imageMetadataRepo.GetImageUserMetadata(ctx, imageURL)
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "failed to get image metadata", err)
	}
	if ownerID != userID && claims.Role != domain.Admin {
		return "", apperr.NewAppError(apperr.ErrForbidden, "user not authorized to access this image", nil)
	}

	downloadURL, err = imageService.imageStore.GetImageDownloadURL(ctx, imageURL)
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "failed to get image download URL", err)
	}
	return downloadURL, nil
}
