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
	projectService services.ProjectService
	authMiddleware *middleware.AuthMiddleware
)

func init() {
	ctx := context.Background()

	cfg := config.LoadConfig()

	ddbClient, err := dynamodb.InitDynamoDBClient(ctx)
	if err != nil {
		panic(err)
	}

	projectRepo := dynamodb.NewProjectRepository(ddbClient, cfg.DYNAMODB_TABLE)
	tokenProvider := jwttoken.NewJWTService(
		cfg.JWT_SECRET,
		cfg.JWT_ISSUER,
		cfg.JWT_AUDIENCE,
	)
	projectService = services.NewProjectService(projectRepo)
	authMiddleware = middleware.NewAuthMiddleware(tokenProvider)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	updateProjectRequest, err := utils.DecodeJson[dtos.UpdateProjectRequest](event.Body)
	if err != nil {
		return utils.BuildErrorResponse(http.StatusBadRequest, "invalid request"), nil
	}

	projectID := event.PathParameters["id"]
	projectDomain := dtos.ToProjectDomain(dtos.Project{
		ID:           projectID,
		Name:         updateProjectRequest.Name,
		Description:  updateProjectRequest.Description,
		Budget:       updateProjectRequest.Budget,
		StartDate:    updateProjectRequest.StartDate,
		EndDate:      updateProjectRequest.EndDate,
		DepartmentID: updateProjectRequest.DepartmentID,
	})

	err = projectService.UpdateProject(ctx, projectDomain)
	if err != nil {
		if apperr.IsInvalidError(err) {
			return utils.BuildErrorResponse(http.StatusBadRequest, "invalid project data"), nil
		} else if apperr.IsNotFoundError(err) {
			return utils.BuildErrorResponse(http.StatusNotFound, "project not found"), nil
		} else {
			return utils.HandleDefaultErrors(err), nil
		}
	}

	return utils.BuildResponse(http.StatusOK, "project updated successfully", struct{}{}), nil
}

func main() {
	lambda.Start(middleware.WithCors(authMiddleware.Authenticated(handler)))
}
