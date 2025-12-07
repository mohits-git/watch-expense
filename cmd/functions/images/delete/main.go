package main

import (
	"context"
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
	var err error
	decodedBody := event.Body
	if event.IsBase64Encoded {
		decodedBody, err = utils.DecodeBase64(event.Body)
		if err != nil {
			return utils.BuildErrorResponse(http.StatusBadRequest, "invalid base64 encoding"), nil
		}
	}
	deleteReq, err := utils.DecodeJson[dtos.DeleteImageRequest](decodedBody)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusBadRequest, "invalid request body"), nil
	}
	url := deleteReq.ImageURL
	err = imageService.DeleteUserImage(ctx, url)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusInternalServerError, "Could not delete the image"), nil
	}
	return utils.BuildResponse(http.StatusOK, "Successfully deleted image.", struct{}{}), nil
}

func main() {
	lambda.Start(middleware.WithCors(authMiddleware.Authenticated(handler)))
}
