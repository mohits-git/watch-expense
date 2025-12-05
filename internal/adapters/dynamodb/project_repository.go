package dynamodb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/mohits-git/watch-expense/internal/domain"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

type ProjectRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewProjectRepository(client *dynamodb.Client, tableName string) ports.ProjectRepository {
	return &ProjectRepository{
		client:    client,
		tableName: tableName,
	}
}

func (repo *ProjectRepository) SaveProject(ctx context.Context, project domain.Project) (string, error) {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	if project.CreatedAt <= 0 {
		project.CreatedAt = time.Now().UnixMilli()
		project.UpdatedAt = project.CreatedAt
	}
	_, err := repo.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName: aws.String(repo.tableName),
					Item: map[string]types.AttributeValue{
						"PK":           &types.AttributeValueMemberS{Value: "PROJECT"},
						"SK":           &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#%s", project.DepartmentID, project.ID)},
						"ProjectID":    &types.AttributeValueMemberS{Value: project.ID},
						"Name":         &types.AttributeValueMemberS{Value: project.Name},
						"Description":  &types.AttributeValueMemberS{Value: project.Description},
						"Budget":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%v", project.Budget)},
						"StartDate":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", project.StartDate)},
						"EndDate":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", project.EndDate)},
						"DepartmentID": &types.AttributeValueMemberS{Value: project.DepartmentID},
						"CreatedAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", project.CreatedAt)},
						"UpdatedAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", project.UpdatedAt)},
					},
				},
			},
			{
				Put: &types.Put{
					TableName: aws.String(repo.tableName),
					Item: map[string]types.AttributeValue{
						"PK":           &types.AttributeValueMemberS{Value: "PROJECT"},
						"SK":           &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPARTMENT#%s", project.ID)},
						"DepartmentID": &types.AttributeValueMemberS{Value: project.DepartmentID},
					},
				},
			},
		},
	})

	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "Error saving project to dynamodb", err)
	}
	return project.ID, nil
}

func (repo *ProjectRepository) UpdateProject(ctx context.Context, project domain.Project) error {
	prevDepartmentId, err := repo.findDepartmentIdByProject(ctx, project.ID)
	if err != nil {
		return err
	}

	project.UpdatedAt = time.Now().UnixMilli()
	updateExpression, err := repo.buildUpdateExpression(project)
	if err != nil {
		return err
	}

	updateTransactionItems := []types.TransactWriteItem{
		{
			Update: &types.Update{
				TableName: aws.String(repo.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "PROJECT"},
					"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#%s", prevDepartmentId, project.ID)},
				},
				UpdateExpression:          updateExpression.Update(),
				ExpressionAttributeNames:  updateExpression.Names(),
				ExpressionAttributeValues: updateExpression.Values(),
			},
		},
	}

	if prevDepartmentId != project.DepartmentID {
		updateTransactionItems = append(updateTransactionItems, types.TransactWriteItem{
			Update: &types.Update{
				TableName: aws.String(repo.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "PROJECT"},
					"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPARTMENT#%s", project.ID)},
				},
				UpdateExpression: aws.String("SET DepartmentID = :departmentId"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":departmentId": &types.AttributeValueMemberS{Value: project.DepartmentID},
				},
			},
		})
	}

	_, err = repo.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: updateTransactionItems,
	})
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "Error while updating project data", err)
	}
	return nil
}

func (repo *ProjectRepository) FindProjectById(ctx context.Context, projectId string) (domain.Project, error) {
	departmentId, err := repo.findDepartmentIdByProject(ctx, projectId)
	if err != nil {
		return domain.Project{}, err
	}
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "PROJECT"},
			":sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#%s", departmentId, projectId)},
		},
	})
	if err != nil {
		return domain.Project{}, err
	}
	if len(result.Items) <= 0 {
		return domain.Project{}, apperr.NewAppError(apperr.ErrNotFound, "project not found with specified id", nil)
	}
	return repo.toDomainProject(result.Items[0]), nil
}

func (repo *ProjectRepository) FindAllProjects(ctx context.Context) ([]domain.Project, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "PROJECT"},
			":sk": &types.AttributeValueMemberS{Value: "DETAILS#"},
		},
	})
	if err != nil {
		return []domain.Project{}, err
	}
	projects := []domain.Project{}
	for _, projectItem := range result.Items {
		projects = append(projects, repo.toDomainProject(projectItem))
	}
	return projects, nil
}

func (repo *ProjectRepository) buildUpdateExpression(project domain.Project) (expression.Expression, error) {
	update := expression.Set(expression.Name("Name"), expression.Value(project.Name))
	update.Set(expression.Name("Description"), expression.Value(project.Description))
	update.Set(expression.Name("Budget"), expression.Value(project.Budget))
	update.Set(expression.Name("DepartmentID"), expression.Value(project.DepartmentID))
	update.Set(expression.Name("StartDate"), expression.Value(project.StartDate))
	update.Set(expression.Name("EndDate"), expression.Value(project.EndDate))
	update.Set(expression.Name("UpdatedAt"), expression.Value(project.UpdatedAt))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return expression.Expression{}, apperr.NewAppError(apperr.ErrInternal, "Error while building update expression", err)
	}
	return expr, nil
}

func (repo *ProjectRepository) findDepartmentIdByProject(ctx context.Context, projectId string) (string, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "PROJECT"},
			":sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPARTMENT#%s", projectId)},
		},
	})
	if err != nil {
		return "", err
	}
	if len(result.Items) == 0 {
		return "", apperr.NewAppError(apperr.ErrNotFound, "Project not found for specified id", nil)
	}
	departmentId := result.Items[0]["DepartmentID"].(*types.AttributeValueMemberS).Value
	return departmentId, nil
}

func (repo *ProjectRepository) toDomainProject(projectItem map[string]types.AttributeValue) domain.Project {
	startDate, _ := strconv.ParseInt(projectItem["StartDate"].(*types.AttributeValueMemberN).Value, 10, 64)
	endDate, _ := strconv.ParseInt(projectItem["EndDate"].(*types.AttributeValueMemberN).Value, 10, 64)
	createdAt, _ := strconv.ParseInt(projectItem["CreatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	updatedAt, _ := strconv.ParseInt(projectItem["UpdatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	budget, _ := strconv.ParseFloat(projectItem["Budget"].(*types.AttributeValueMemberN).Value, 64)
	departmentId := ""
	if projectItem["DepartmentID"] != nil {
		departmentAttr, ok := projectItem["DepartmentID"].(*types.AttributeValueMemberS)
		if ok {
			departmentId = departmentAttr.Value
		}
	}
	return domain.Project{
		ID:           projectItem["ProjectID"].(*types.AttributeValueMemberS).Value,
		Name:         projectItem["Name"].(*types.AttributeValueMemberS).Value,
		Description:  projectItem["Description"].(*types.AttributeValueMemberS).Value,
		Budget:       budget,
		DepartmentID: departmentId,
		StartDate:    startDate,
		EndDate:      endDate,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
