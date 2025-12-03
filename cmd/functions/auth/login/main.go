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
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/utils"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

var (
	authService services.AuthenticationService
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
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	loginRequest, err := utils.DecodeJson[dtos.LoginRequest](event.Body)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusBadRequest, "invalid request"), nil
	}

	token, err := authService.Login(ctx, loginRequest.Email, loginRequest.Password)
	if err != nil {
		if apperr.IsNotFoundError(err) || apperr.IsUnauthorizedError(err) {
			return utils.BuildErrorResponse(http.StatusUnauthorized, "invalid email or password"), nil
		} else if apperr.IsInvalidError(err) {
			return utils.BuildErrorResponse(http.StatusBadRequest, "invalid inputs"), nil
		} else {
			return utils.BuildErrorResponse(http.StatusInternalServerError, "internal server error"), nil
		}
	}

	loginResponse := dtos.LoginResponse{Token: token}
	resp := utils.BuildResponse(http.StatusOK, "login successful", loginResponse)
	resp.Headers["Set-Cookie"] = "token=" + token + "; HttpOnly; Path=/api/; SameSite=Strict"
	return resp, nil
}

func main() {
	lambda.Start(handler)
}
