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
	advanceService services.AdvanceService
	authMiddleware *middleware.AuthMiddleware
)

func init() {
	ctx := context.Background()

	cfg := config.LoadConfig()

	ddbClient, err := dynamodb.InitDynamoDBClient(ctx)
	if err != nil {
		panic(err)
	}

	advanceRepo := dynamodb.NewAdvanceRepository(ddbClient, cfg.DYNAMODB_TABLE)
	tokenProvider := jwttoken.NewJWTService(
		cfg.JWT_SECRET,
		cfg.JWT_ISSUER,
		cfg.JWT_AUDIENCE,
	)
	advanceService = services.NewAdvanceService(advanceRepo)
	authMiddleware = middleware.NewAuthMiddleware(tokenProvider)
}

func parseAdvancesFilterOptions(event events.APIGatewayProxyRequest) (domain.AdvancesFilterOptions, error) {
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
			return domain.AdvancesFilterOptions{}, err
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return domain.AdvancesFilterOptions{}, err
		}
	}

	return domain.AdvancesFilterOptions{
		UserID: userID,
		Page:   page,
		Limit:  limit,
		Status: domain.RequestStatus(status),
	}, nil
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	filterOptions, err := parseAdvancesFilterOptions(event)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusBadRequest, "invalid query params"), nil
	}

	advances, total, err := advanceService.GetAllAdvances(ctx, filterOptions)
	if err != nil {
		if apperr.IsInvalidError(err) {
			return utils.BuildErrorResponse(http.StatusBadRequest, "invalid request parameters"), nil
		}
		return utils.HandleDefaultErrors(err), nil
	}
	advanceResponse := []dtos.Advance{}
	for _, advance := range advances {
		advanceResponse = append(advanceResponse, dtos.ToAdvanceDTO(advance))
	}
	getAdvancesResponse := dtos.GetAdvancesResponse{
		TotalAdvances: total,
		Advances:      advanceResponse,
	}
	return utils.BuildResponse(http.StatusOK, "advances fetched successfully", getAdvancesResponse), nil
}

func main() {
	lambda.Start(middleware.WithCors(authMiddleware.Authenticated(handler)))
}
