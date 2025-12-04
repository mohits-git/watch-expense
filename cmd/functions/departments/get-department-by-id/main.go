package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mohits-git/watch-expense/internal/adapters/dynamodb"
	"github.com/mohits-git/watch-expense/internal/adapters/http/dtos"
	"github.com/mohits-git/watch-expense/internal/adapters/jwttoken"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/config"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/middleware"
	"github.com/mohits-git/watch-expense/internal/adapters/lambda/utils"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

var (
	departmentService services.DepartmentService
	authMiddleware    *middleware.AuthMiddleware
)

func init() {
	ctx := context.Background()

	cfg := config.LoadConfig()

	ddbClient, err := dynamodb.InitDynamoDBClient(ctx)
	if err != nil {
		panic(err)
	}

	departmentRepo := dynamodb.NewDepartmentRepository(ddbClient, cfg.DYNAMODB_TABLE)
	tokenProvider := jwttoken.NewJWTService(
		cfg.JWT_SECRET,
		cfg.JWT_ISSUER,
		cfg.JWT_AUDIENCE,
	)
	departmentService = services.NewDepartmentService(departmentRepo)
	authMiddleware = middleware.NewAuthMiddleware(tokenProvider)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	departmentID := event.PathParameters["id"]
	department, err := departmentService.GetDepartmentByID(ctx, departmentID)
	if err != nil {
		if apperr.IsNotFoundError(err) {
			return utils.BuildErrorResponse(http.StatusNotFound, "department not found"), nil
		} else if apperr.IsInvalidError(err) {
			return utils.BuildErrorResponse(http.StatusBadRequest, "invalid department ID"), nil
		} else {
			return utils.HandleDefaultErrors(err), nil
		}
	}
	responseDepartment := dtos.ToDepartmentDTO(department)
	return utils.BuildResponse(http.StatusOK, "department fetched successfully", responseDepartment), nil
}

func main() {
	lambda.Start(middleware.WithCors(authMiddleware.Authenticated(handler)))
}
