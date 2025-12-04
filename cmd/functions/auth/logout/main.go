package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mohits-git/watch-expense/internal/adapters/bcrypt"
	"github.com/mohits-git/watch-expense/internal/adapters/dynamodb"
	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/adapters/jwttoken"
	cfg "github.com/mohits-git/watch-expense/internal/adapters/lambda/config"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/middleware"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/utils"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/authctx"
)

var (
	authService    services.AuthenticationService
	authMiddleware *middleware.AuthMiddleware
)

func init() {
	ctx := context.Background()

	cfg := cfg.LoadConfig()

	ddbClient, err := dynamodb.InitDynamoDBClient(ctx)
	if err != nil {
		panic(err)
	}

	userRepo := dynamodb.NewUserRepository(ddbClient, cfg.DYNAMODB_TABLE)
	bcryptProvider := bcrypt.NewBcryptPasswordHasher(12)
	tokenProvider := jwttoken.NewJWTService(
		cfg.JWT_SECRET,
		cfg.JWT_ISSUER,
		cfg.JWT_AUDIENCE,
	)

	authService = services.NewAuthenticationService(userRepo, tokenProvider, bcryptProvider)
	authMiddleware = middleware.NewAuthMiddleware(tokenProvider)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	token, ok := authctx.TokenFromCtx(ctx)
	if !ok {
		return utils.BuildErrorResponse(http.StatusUnauthorized, "unauthorized"), nil
	}

	err := authService.Logout(ctx, token)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusInternalServerError, "internal server error"), nil
	}

	logoutResponse := dtos.LogoutResponse{}
	resp := utils.BuildResponse(http.StatusOK, "logout successful", logoutResponse)
	// resp.Headers["Set-Cookie"] = "token=; HttpOnly; Path=/api/; Max-Age=0; SameSite=Strict"
	return resp, nil
}

func main() {
	lambda.Start(middleware.WithCors(authMiddleware.WithToken(handler)))
}
