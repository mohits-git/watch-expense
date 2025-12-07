package main

import (
	"context"
	"net/http"
	"strconv"

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

func parseExpensesFilterOptions(event events.APIGatewayProxyRequest) (domain.ExpensesFilterOptions, error) {
	var err error
	status := event.QueryStringParameters["status"]
	pageStr := event.QueryStringParameters["page"]
	page := 1
	limitStr := event.QueryStringParameters["limit"]
	limit := 10
	userID := event.QueryStringParameters["user_id"]

	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return domain.ExpensesFilterOptions{}, err
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return domain.ExpensesFilterOptions{}, err
		}
	}

	return domain.ExpensesFilterOptions{
		UserID: userID,
		Page:   page,
		Limit:  limit,
		Status: domain.RequestStatus(status),
	}, nil
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filterOptions, err := parseExpensesFilterOptions(event)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusBadRequest, "invalid query params"), nil
	}

	expenses, total, err := expenseService.GetAllExpenses(ctx, filterOptions)
	if err != nil {
		if apperr.IsInvalidError(err) {
			return utils.BuildErrorResponse(http.StatusBadRequest, "invalid request parameters"), nil
		}
		return utils.HandleDefaultErrors(err), nil
	}
	expenseResponse := []dtos.Expense{}
	for _, expense := range expenses {
		expenseResponse = append(expenseResponse, dtos.ToExpenseDTO(expense))
	}
	getExpensesResponse := dtos.GetExpensesResponse{
		TotalExpenses: total,
		Expenses:      expenseResponse,
	}
	return utils.BuildResponse(http.StatusOK, "expenses fetched successfully", getExpensesResponse), nil
}

func main() {
	lambda.Start(middleware.WithCors(authMiddleware.Authenticated(handler)))
}
