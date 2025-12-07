package main

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mohits-git/watch-expense/internal/adapters/dynamodb"
	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/adapters/jwttoken"
	cfg "github.com/mohits-git/watch-expense/internal/adapters/lambda/config"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/middleware"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/utils"
	"github.com/mohits-git/watch-expense/internal/adapters/s3imagestore"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

var (
	imageService   services.ImageService
	authMiddleware *middleware.AuthMiddleware
)

func init() {
	ctx := context.Background()

	cfg := cfg.LoadConfig()

	s3ImageStore, err := s3imagestore.NewS3ImageStore(ctx, cfg.S3_BUCKET_NAME)
	if err != nil {
		panic(err)
	}

	ddbclient, err := dynamodb.InitDynamoDBClient(ctx)
	if err != nil {
		panic(err)
	}
	imageMetadataRepo := dynamodb.NewImageMetadataRepository(ddbclient, cfg.DYNAMODB_TABLE)

	tokenProvider := jwttoken.NewJWTService(
		cfg.JWT_SECRET,
		cfg.JWT_ISSUER,
		cfg.JWT_AUDIENCE,
	)

	imageService = services.NewImageService(s3ImageStore, imageMetadataRepo)
	authMiddleware = middleware.NewAuthMiddleware(tokenProvider)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	decodedBody, err := utils.DecodeBase64(event.Body)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusInternalServerError, "internal server error"), nil
	}

	contentType := event.Headers["Content-Type"]
	if contentType == "" {
		contentType = event.Headers["content-type"]
	}
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusInternalServerError, "internal server error"), nil
	}
	boundary := params["boundary"]

	reader := multipart.NewReader(bytes.NewReader([]byte(decodedBody)), boundary)

	file, err := reader.NextPart()
	if err != nil && err != io.EOF {
		return utils.BuildErrorResponse(http.StatusInternalServerError, "internal server error"), nil
	}
	fileName := file.FileName()
	if fileName == "" {
		return utils.BuildErrorResponse(http.StatusBadRequest, "invalid data"), nil
	}
	defer file.Close()

	url, err := imageService.UploadUserImage(ctx, file, fileName)
	if err != nil {
		if apperr.IsTooLargeError(err) {
			return utils.BuildErrorResponse(413, "File too large"), nil
		}
		return utils.BuildErrorResponse(http.StatusInternalServerError, "Something went wrong while uploading your image."), nil
	}
	response := dtos.ImageUploadResponse{ImageURL: url}
	return utils.BuildResponse(http.StatusCreated, "Successfully uploaded image.", response), nil
}

func main() {
	lambda.Start(middleware.WithCors(authMiddleware.Authenticated(handler)))
}
