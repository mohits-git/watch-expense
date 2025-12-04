package middleware

import (
	"context"
	"maps"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/utils"
)

func WithCors(handlerFn utils.LambdaHandlerFunction) utils.LambdaHandlerFunction {
	return utils.LambdaHandlerFunction(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		if req.HTTPMethod == "OPTIONS" {
			return events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers:    addCorsHeaders(nil),
			}, nil
		}
		res, err := handlerFn(ctx, req)
		res.Headers = addCorsHeaders(res.Headers)
		return res, err
	})
}

func addCorsHeaders(headers map[string]string) map[string]string {
	withCorsHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token",
		"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS,PATCH",
	}

	if headers != nil {
    maps.Copy(withCorsHeaders, headers)
	}

	return withCorsHeaders
}
