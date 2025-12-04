package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

func EncodeJson[T any](data T) (string, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "failed to encode json response", err)
	}
	return string(bytes), nil
}

func DecodeJson[T any](body string) (T, error) {
	var data T
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		return data, apperr.NewAppError(apperr.ErrInvalid, "failed to decode json", err)
	}
	return data, nil
}

func BuildResponse[T any](statusCode int, msg string, data T) events.APIGatewayProxyResponse {
	body, _ := EncodeJson(dtos.NewResponse(statusCode, msg, data))
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: body,
	}
}

func BuildErrorResponse(statusCode int, msg string) events.APIGatewayProxyResponse {
	return BuildResponse(statusCode, msg, struct{}{})
}

func GetBearerToken(event events.APIGatewayProxyRequest) string {
	authHeader := event.Headers["Authorization"]
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

type LambdaHandlerFunction = func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

func HandleDefaultErrors(err error) events.APIGatewayProxyResponse {
	switch {
	case apperr.IsUnauthorizedError(err):
		return BuildErrorResponse(http.StatusUnauthorized, "unauthorized")
	case apperr.IsForbiddenError(err):
		return BuildErrorResponse(http.StatusForbidden, "forbidden")
	case apperr.IsInternalError(err):
		return BuildErrorResponse(http.StatusInternalServerError, "internal server error")
	}
	return BuildErrorResponse(http.StatusInternalServerError, "internal server error")
}
