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

type UserRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewUserRepository(client *dynamodb.Client, tableName string) ports.UserRepository {
	return &UserRepository{
		client:    client,
		tableName: tableName,
	}
}

func (repo *UserRepository) SaveUser(ctx context.Context, user domain.User) (string, error) {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt <= 0 {
		user.CreatedAt = time.Now().UnixMilli()
		user.UpdatedAt = user.CreatedAt
	}
	_, err := repo.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName: aws.String(repo.tableName),
					Item: map[string]types.AttributeValue{
						"PK":           &types.AttributeValueMemberS{Value: "USER"},
						"SK":           &types.AttributeValueMemberS{Value: fmt.Sprintf("PROFILE#%s", user.ID)},
						"UserID":       &types.AttributeValueMemberS{Value: user.ID},
						"EmployeeID":   &types.AttributeValueMemberS{Value: user.EmployeeId},
						"Name":         &types.AttributeValueMemberS{Value: user.Name},
						"Email":        &types.AttributeValueMemberS{Value: user.Email},
						"Role":         &types.AttributeValueMemberS{Value: string(user.Role)},
						"PasswordHash": &types.AttributeValueMemberS{Value: user.Password},
						"DepartmentID": &types.AttributeValueMemberS{Value: user.DepartmentID},
						"ProjectID":    &types.AttributeValueMemberS{Value: user.ProjectID},
						"CreatedAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", user.CreatedAt)},
						"UpdatedAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", user.UpdatedAt)},
					},
				},
			},
			{
				Put: &types.Put{
					TableName: aws.String(repo.tableName),
					Item: map[string]types.AttributeValue{
						"PK":     &types.AttributeValueMemberS{Value: "USER"},
						"SK":     &types.AttributeValueMemberS{Value: fmt.Sprintf("EMAIL#%s", user.Email)},
						"UserID": &types.AttributeValueMemberS{Value: user.ID},
					},
				},
			},
		},
	})
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "Error saving user to dynamodb", err)
	}
	return user.ID, nil
}

func (repo *UserRepository) buildUpdateExpression(user domain.User) (expression.Expression, error) {
	update := expression.Set(expression.Name("EmployeeID"), expression.Value(user.EmployeeId))
	update.Set(expression.Name("Name"), expression.Value(user.Name))
	update.Set(expression.Name("Email"), expression.Value(user.Email))
	update.Set(expression.Name("Role"), expression.Value(user.Role))
	if user.Password != "" {
		update.Set(expression.Name("PasswordHash"), expression.Value(user.Password))
	}
	update.Set(expression.Name("DepartmentID"), expression.Value(user.DepartmentID))
	update.Set(expression.Name("ProjectID"), expression.Value(user.ProjectID))
	update.Set(expression.Name("UpdatedAt"), expression.Value(user.UpdatedAt))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return expression.Expression{}, apperr.NewAppError(apperr.ErrInternal, "Error while building update expression", err)
	}
	return expr, nil
}

func (repo *UserRepository) UpdateUser(ctx context.Context, user domain.User) error {
	userExist, err := repo.FindUserById(ctx, user.ID)
	if err != nil {
		return err
	}

	user.UpdatedAt = time.Now().UnixMilli()
	updateExpression, err := repo.buildUpdateExpression(user)
	if err != nil {
		return err
	}

	updateTransactionItems := []types.TransactWriteItem{
		{
			Update: &types.Update{
				TableName: aws.String(repo.tableName),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER"},
					"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("PROFILE#%s", user.ID)},
				},
				UpdateExpression:          updateExpression.Update(),
				ExpressionAttributeNames:  updateExpression.Names(),
				ExpressionAttributeValues: updateExpression.Values(),
			},
		},
	}

	if userExist.Email != user.Email {
		updateTransactionItems = append(
			updateTransactionItems,
			types.TransactWriteItem{
				Delete: &types.Delete{
					TableName: aws.String(repo.tableName),
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{Value: "USER"},
						"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMAIL#%s", userExist.Email)},
					},
				},
			},
			types.TransactWriteItem{
				Put: &types.Put{
					TableName: aws.String(repo.tableName),
					Item: map[string]types.AttributeValue{
						"PK":     &types.AttributeValueMemberS{Value: "USER"},
						"SK":     &types.AttributeValueMemberS{Value: fmt.Sprintf("EMAIL#%s", user.Email)},
						"UserID": &types.AttributeValueMemberS{Value: user.ID},
					},
				},
			},
		)
	}

	_, err = repo.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: updateTransactionItems,
	})
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "Error while updating user data", err)
	}
	return nil
}

func (repo *UserRepository) FindUserById(ctx context.Context, userId string) (domain.User, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER"},
			":sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("PROFILE#%s", userId)},
		},
	})
	if err != nil {
		return domain.User{}, err
	}
	if len(result.Items) == 0 {
		return domain.User{}, apperr.NewAppError(apperr.ErrNotFound, "User not found for specified id", nil)
	}
	userItem := result.Items[0]
	user := repo.toDomainUser(userItem)
	return user, nil
}

func (repo *UserRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	userId, err := repo.findUserIdByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	user, err := repo.FindUserById(ctx, userId)
	if err != nil {
		return domain.User{}, apperr.NewAppError(apperr.ErrInternal, "Something went wrong while fetching user", nil)
	}
	return user, nil
}

func (repo *UserRepository) FindAllUsers(ctx context.Context) ([]domain.User, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER"},
			":sk": &types.AttributeValueMemberS{Value: "PROFILE#"},
		},
	})
	if err != nil {
		return []domain.User{}, err
	}
	users := []domain.User{}
	for _, userItem := range result.Items {
		users = append(users, repo.toDomainUser(userItem))
	}
	return users, nil
}

func (repo *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	userExist, err := repo.FindUserById(ctx, userID)
	if err != nil {
		return err
	}
	_, err = repo.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Delete: &types.Delete{
					TableName: aws.String(repo.tableName),
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{Value: "USER"},
						"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("PROFILE#%s", userID)},
					},
				},
			},
			{
				Delete: &types.Delete{
					TableName: aws.String(repo.tableName),
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{Value: "USER"},
						"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMAIL#%s", userExist.Email)},
					},
				},
			},
		},
	})
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "Error while deleting user", err)
	}
	return nil
}

func (repo *UserRepository) findUserIdByEmail(ctx context.Context, email string) (string, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER"},
			":sk": &types.AttributeValueMemberS{Value: fmt.Sprintf("EMAIL#%s", email)},
		},
	})
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "Something went wrong", err)
	}
	if len(result.Items) == 0 {
		return "", apperr.NewAppError(apperr.ErrNotFound, "User not found for specified email", nil)
	}
	userAttr, ok := result.Items[0]["UserID"].(*types.AttributeValueMemberS)
	if !ok {
		return "", apperr.NewAppError(apperr.ErrInternal, "Something went wrong", err)
	}
	return userAttr.Value, nil
}

func (repo *UserRepository) toDomainUser(userItem map[string]types.AttributeValue) domain.User {
	createdAt, _ := strconv.ParseInt(userItem["CreatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	updatedAt, _ := strconv.ParseInt(userItem["UpdatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	departmentId, projectId := "", ""
	if userItem["DepartmentID"] != nil {
		departmentAttr, ok := userItem["DepartmentID"].(*types.AttributeValueMemberS)
		if ok {
			departmentId = departmentAttr.Value
		}
	}
	if userItem["ProjectID"] != nil {
		projectAttr, ok := userItem["ProjectID"].(*types.AttributeValueMemberS)
		if ok {
			projectId = projectAttr.Value
		}

	}
	return domain.User{
		ID:           userItem["UserID"].(*types.AttributeValueMemberS).Value,
		Name:         userItem["Name"].(*types.AttributeValueMemberS).Value,
		Email:        userItem["Email"].(*types.AttributeValueMemberS).Value,
		EmployeeId:   userItem["EmployeeID"].(*types.AttributeValueMemberS).Value,
		Role:         domain.UserRole(userItem["Role"].(*types.AttributeValueMemberS).Value),
		Password:     userItem["PasswordHash"].(*types.AttributeValueMemberS).Value,
		DepartmentID: departmentId,
		ProjectID:    projectId,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
}
