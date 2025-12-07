package s3imagestore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

type S3ImageStore struct {
	client          *s3.Client
	presignedClient *s3.PresignClient
	bucketName      string
}

func NewS3ImageStore(ctx context.Context, bucket string) (ports.ImageStore, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg)
	presignedClient := s3.NewPresignClient(s3Client)

	return &S3ImageStore{
		client:          s3Client,
		presignedClient: presignedClient,
		bucketName:      bucket,
	}, nil
}

func (s3IU *S3ImageStore) UploadImage(ctx context.Context, imageData io.Reader, name string) (url string, err error) {
	objectKey := uuid.New().String() + "_" + name

	imageBytes, err := io.ReadAll(imageData)
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInvalid, "unable to read image data", err)
	}
	contentLength := int64(len(imageBytes))

	_, err = s3IU.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s3IU.bucketName),
		Key:           aws.String(objectKey),
		Body:          bytes.NewBuffer(imageBytes),
		ContentLength: aws.Int64(contentLength),
	})

	var apiErr smithy.APIError
	if err != nil {
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			return "", apperr.NewAppError(apperr.ErrTooLarge, "image size exceeds the allowed limit", err)
		}
		return "", apperr.NewAppError(apperr.ErrInternal, "failed to upload image", err)
	}

	// wait until the object is uploaded
	err = s3.NewObjectExistsWaiter(s3IU.client).Wait(
		ctx,
		&s3.HeadObjectInput{Bucket: aws.String(s3IU.bucketName), Key: aws.String(objectKey)},
		time.Minute,
	)
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "failed to confirm image upload", err)
	}

	url = "https://" + s3IU.bucketName + ".s3.amazonaws.com/" + objectKey
	return url, nil
}

func (s3IU *S3ImageStore) DeleteImage(ctx context.Context, imageUrl string) error {
	objectKey := imageUrl[strings.LastIndex(imageUrl, "/")+1:]
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s3IU.bucketName),
		Key:    aws.String(objectKey),
	}

	_, err := s3IU.client.DeleteObject(ctx, input)
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return apperr.NewAppError(apperr.ErrNotFound, "image not found", err)
		}
		return apperr.NewAppError(apperr.ErrInternal, "failed to delete image", err)
	}

	err = s3.NewObjectNotExistsWaiter(s3IU.client).Wait(
		ctx,
		&s3.HeadObjectInput{Bucket: aws.String(s3IU.bucketName), Key: aws.String(objectKey)},
		10*time.Second,
	)
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "failed to confirm image deletion", err)
	}

	return nil
}

func (s3IU *S3ImageStore) GetImageDownloadURL(ctx context.Context, objectUrl string) (string, error) {
	objectKey := objectUrl[strings.LastIndex(objectUrl, "/")+1:]
	var lifetimeSecs int64 = 60
	request, err := s3IU.presignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3IU.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "failed to generate presigned URL", err)
	}
	return request.URL, nil
}
