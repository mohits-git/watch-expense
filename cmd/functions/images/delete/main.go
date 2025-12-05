package main

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/adapters/imageupload"
	cfg "github.com/mohits-git/watch-expense/internal/adapters/lambda/config"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/middleware"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/utils"
	"github.com/mohits-git/watch-expense/internal/ports"
)

var (
	imageUploadService ports.ImageUploadService
)

func init() {
	ctx := context.Background()

	cfg := cfg.LoadConfig()

	imageService, err := imageupload.NewS3ImageUpload(ctx, cfg.S3_BUCKET_NAME)
	if err != nil {
		panic(err)
	}

	imageUploadService = imageService
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
	err = imageUploadService.DeleteImage(ctx, url)
	if err != nil {
		log.Println("image delete error: ", err)
		return utils.BuildErrorResponse(http.StatusInternalServerError, "Could not delete the image"), nil
	}
	return utils.BuildResponse(http.StatusOK, "Successfully deleted image.", struct{}{}), nil
}

func main() {
	lambda.Start(middleware.WithCors(handler))
}
