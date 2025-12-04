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
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/services"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

var (
	expenseService services.ExpenseService
	authMiddleware *middleware.AuthMiddleware
)

func init() {
	ctx := context.Background()

	cfg := config.LoadConfig()

	ddbClient, err := dynamodb.InitDynamoDBClient(ctx)
	if err != nil {
		panic(err)
	}

	expenseRepo := dynamodb.NewExpenseRepository(ddbClient, cfg.DYNAMODB_TABLE)
	tokenProvider := jwttoken.NewJWTService(
		cfg.JWT_SECRET,
		cfg.JWT_ISSUER,
		cfg.JWT_AUDIENCE,
	)
	expenseService = services.NewExpenseService(expenseRepo)
	authMiddleware = middleware.NewAuthMiddleware(tokenProvider)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	updateExpenseRequest, err := utils.DecodeJson[dtos.UpdateExpenseRequest](event.Body)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusBadRequest, "invalid request"), nil
	}

	expenseID := event.PathParameters["id"]
	expense := domain.Expense{
		ID:           expenseID,
		Amount:       updateExpenseRequest.Amount,
		Description:  updateExpenseRequest.Description,
		Purpose:      updateExpenseRequest.Purpose,
		IsReconciled: updateExpenseRequest.IsReconciled,
	}

	err = expenseService.UpdateExpense(ctx, expense)
	if err != nil {
		if apperr.IsInvalidError(err) {
			return utils.BuildErrorResponse(http.StatusBadRequest, "invalid expense data"), nil
		} else if apperr.IsNotFoundError(err) {
			return utils.BuildErrorResponse(http.StatusNotFound, "expense not found"), nil
		} else {
			return utils.HandleDefaultErrors(err), nil
		}
	}

	return utils.BuildResponse(http.StatusOK, "expense updated successfully", struct{}{}), nil
}

func main() {
	lambda.Start(middleware.WithCors(authMiddleware.Authenticated(handler)))
}
