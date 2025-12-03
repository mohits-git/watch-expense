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
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

var (
	userService    services.UserService
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

	userService = services.NewUserService(userRepo, nil, bcryptProvider)
	authMiddleware = middleware.NewAuthMiddleware(tokenProvider)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userId := event.PathParameters["id"]
	updateUserRequest, err := utils.DecodeJson[dtos.UpdateUserRequest](event.Body)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusBadRequest, "invalid request"), nil
	}

	user := domain.User{
		ID:           userId,
		EmployeeId:   updateUserRequest.EmployeeId,
		Name:         updateUserRequest.Name,
		Password:     updateUserRequest.Password,
		Email:        updateUserRequest.Email,
		Role:         updateUserRequest.Role,
		ProjectID:    updateUserRequest.ProjectID,
		DepartmentID: updateUserRequest.DepartmentID,
	}

	err = userService.UpdateUser(ctx, user)
	if err != nil {
		if apperr.IsInvalidError(err) {
			return utils.BuildErrorResponse(http.StatusBadRequest, "invalid user data"), nil
		} else {
			return utils.HandleDefaultErrors(err), nil
		}
	}

	resp := utils.BuildResponse(http.StatusOK, "user updated successfully", struct{}{})
	return resp, nil
}

func main() {
	lambda.Start(authMiddleware.Authenticated(handler))
}
