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
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
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
	createUserRequest, err := utils.DecodeJson[dtos.CreateUserRequest](event.Body)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusBadRequest, "invalid request"), nil
	}

	user := dtos.ToUserDomain(dtos.User{
		EmployeeId:   createUserRequest.EmployeeId,
		Name:         createUserRequest.Name,
		Password:     createUserRequest.Password,
		Email:        createUserRequest.Email,
		Role:         createUserRequest.Role,
		ProjectID:    createUserRequest.ProjectID,
		DepartmentID: createUserRequest.DepartmentID,
	})
	userID, err := userService.CreateUser(ctx, user)
	if err != nil {
		if apperr.IsInvalidError(err) {
			return utils.BuildErrorResponse(http.StatusBadRequest, "invalid user data"), nil
		} else {
			return utils.HandleDefaultErrors(err), err
		}
	}

	resp := utils.BuildResponse(http.StatusCreated, "user created successfully", dtos.CreateUserResponse{ID: userID})
	return resp, nil
}

func main() {
	lambda.Start(authMiddleware.Authenticated(handler))
}
