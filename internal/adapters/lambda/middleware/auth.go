package middleware

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/utils"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
)

type AuthMiddleware struct {
	tokenProvider ports.TokenProvider
}

func NewAuthMiddleware(tokenProvider ports.TokenProvider) *AuthMiddleware {
	return &AuthMiddleware{
		tokenProvider: tokenProvider,
	}
}

func (m *AuthMiddleware) Authenticated(handler utils.LambdaHanlderFunction) utils.LambdaHanlderFunction {
	return utils.LambdaHanlderFunction(func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		token := utils.GetBearerToken(event)
		if token == "" {
			return utils.BuildErrorResponse(http.StatusUnauthorized, "missing or invalid token"), nil
		}

		// validate the token
		userClaims, err := m.tokenProvider.ValidateToken(token)
		if err != nil {
			return utils.BuildErrorResponse(http.StatusUnauthorized, "invalid token"), nil
		}

		authctx := authctx.WithUserClaims(ctx, &userClaims)
		return handler(authctx, event)
	})
}
