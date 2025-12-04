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

type DepartmentRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDepartmentRepository(client *dynamodb.Client, tableName string) ports.DepartmentRepository {
	return &DepartmentRepository{
		client:    client,
		tableName: tableName,
	}
}

func (repo *DepartmentRepository) SaveDepartment(ctx context.Context, department domain.Department) (string, error) {
	if department.ID == "" {
		department.ID = uuid.New().String()
	}
	if department.CreatedAt <= 0 {
		department.CreatedAt = time.Now().UnixMilli()
		department.UpdatedAt = department.CreatedAt
	}

	_, err := repo.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(repo.tableName),
		Item: map[string]types.AttributeValue{
			"PK":           &types.AttributeValueMemberS{Value: "DEPARTMENT"},
			"SK":           &types.AttributeValueMemberS{Value: fmt.Sprintf("DEPARTMENT#%s", department.ID)},
			"DepartmentID": &types.AttributeValueMemberS{Value: department.ID},
			"Name":         &types.AttributeValueMemberS{Value: department.Name},
			"Budget":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%2f", department.Budget)},
			"CreatedAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", department.CreatedAt)},
			"UpdatedAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", department.UpdatedAt)},
		},
	})
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "Error saving department to dynamodb", err)
	}
	return department.ID, nil
}

func (repo *DepartmentRepository) buildUpdateExpression(department domain.Department) (expression.Expression, error) {
	update := expression.Set(expression.Name("Name"), expression.Value(department.Name))
	update.Set(expression.Name("Budget"), expression.Value(department.Budget))
	update.Set(expression.Name("UpdatedAt"), expression.Value(department.UpdatedAt))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return expression.Expression{}, apperr.NewAppError(apperr.ErrInternal, "Error while building update expression", err)
	}
	return expr, nil
}

func (repo *DepartmentRepository) UpdateDepartment(ctx context.Context, department domain.Department) error {
	updateExpression, err := repo.buildUpdateExpression(department)
	if err != nil {
		return err
	}
	_, err = repo.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "DEPARTMENT"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s", department.ID)},
		},
		UpdateExpression:          updateExpression.Update(),
		ExpressionAttributeNames:  updateExpression.Names(),
		ExpressionAttributeValues: updateExpression.Values(),
	})
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "Error while updating department data", err)
	}
	return nil
}

func (repo *DepartmentRepository) FindDepartmentById(ctx context.Context, departmentId string) (domain.Department, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "DEPARTMENT"},
			":sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s", departmentId)},
		},
	})
	if err != nil {
		return domain.Department{}, err
	}
	if len(result.Items) == 0 {
		return domain.Department{}, apperr.NewAppError(apperr.ErrNotFound, "department not found for specified id", nil)
	}
	departmentItem := result.Items[0]
	department := repo.toDomainDepartment(departmentItem)
	return department, nil
}

func (repo *DepartmentRepository) FindAllDepartments(ctx context.Context) ([]domain.Department, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "DEPARTMENT"},
			":sk": &types.AttributeValueMemberS{Value: "DETAILS#"},
		},
	})
	if err != nil {
		return []domain.Department{}, err
	}
	departments := []domain.Department{}
	for _, item := range result.Items {
		departments = append(departments, repo.toDomainDepartment(item))
	}
	return departments, nil
}

func (repo *DepartmentRepository) toDomainDepartment(userItem map[string]types.AttributeValue) domain.Department {
	createdAt, _ := strconv.ParseInt(userItem["CreatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	updatedAt, _ := strconv.ParseInt(userItem["UpdatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	budget, _ := strconv.ParseFloat(userItem["Budget"].(*types.AttributeValueMemberN).Value, 64)
	return domain.Department{
		ID:        userItem["DepartmentID"].(*types.AttributeValueMemberS).Value,
		Name:      userItem["Name"].(*types.AttributeValueMemberS).Value,
		Budget:    budget,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
