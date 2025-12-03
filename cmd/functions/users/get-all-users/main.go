package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	ddb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/mohits-git/watch-expense/internal/adapters/bcrypt"
	"github.com/mohits-git/watch-expense/internal/adapters/dynamodb"
	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/adapters/jwttoken"
	cfg "github.com/mohits-git/watch-expense/internal/adapters/lambda/config"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/middleware"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/utils"
	"github.com/mohits-git/watch-expense/internal/services"
)

var (
	userService    services.UserService
	authMiddleware *middleware.AuthMiddleware
)

func initDynamoDBClient(ctx context.Context) (*ddb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return ddb.NewFromConfig(cfg), nil
}

func init() {
	ctx := context.Background()

	cfg := cfg.LoadConfig()

	ddbClient, err := initDynamoDBClient(ctx)
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

	userService = services.NewUserService(userRepo, nil, bcryptProvider)
	authMiddleware = middleware.NewAuthMiddleware(tokenProvider)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	users, err := userService.GetAllUsers(ctx)
	if err != nil {
		return utils.HandleDefaultErrors(err), nil
	}
	responseUsers := []dtos.User{}
	for _, user := range users {
		responseUsers = append(responseUsers, dtos.ToUserDTO(user))
	}
	resp := utils.BuildResponse(http.StatusOK, "login successful", responseUsers)
	return resp, nil
}

func main() {
	lambda.Start(authMiddleware.Authenticated(handler))
}
