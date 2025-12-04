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

type ExpenseRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewExpenseRepository(client *dynamodb.Client, tableName string) ports.ExpenseRepository {
	return &ExpenseRepository{
		client:    client,
		tableName: tableName,
	}
}

func (repo *ExpenseRepository) SaveExpense(ctx context.Context, expense domain.Expense) (string, error) {
	if expense.ID == "" {
		expense.ID = uuid.New().String()
	}
	if expense.CreatedAt <= 0 {
		expense.CreatedAt = time.Now().UnixMilli()
		expense.UpdatedAt = expense.CreatedAt
	}
	bills := repo.getBillsAttributeValue(expense.Bills)
	_, err := repo.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName: aws.String(repo.tableName),
					Item: map[string]types.AttributeValue{
						"PK":           &types.AttributeValueMemberS{Value: "EXPENSE"},
						"SK":           &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#%s", expense.UserID, expense.ID)},
						"ExpenseID":    &types.AttributeValueMemberS{Value: expense.ID},
						"UserID":       &types.AttributeValueMemberS{Value: expense.UserID},
						"Amount":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", expense.Amount)},
						"Purpose":      &types.AttributeValueMemberS{Value: expense.Purpose},
						"Description":  &types.AttributeValueMemberS{Value: string(expense.Description)},
						"Status":       &types.AttributeValueMemberS{Value: string(expense.Status)},
						"IsReconciled": &types.AttributeValueMemberBOOL{Value: expense.IsReconciled},
						"ApprovedBy":   &types.AttributeValueMemberS{Value: expense.ApprovedBy},
						"ApprovedAt":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expense.ApprovedAt)},
						"ReviewedBy":   &types.AttributeValueMemberS{Value: expense.ReviewedBy},
						"ReviewedAt":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expense.ReviewedAt)},
						"Bills":        bills,
						"CreatedAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expense.CreatedAt)},
						"UpdatedAt":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", expense.UpdatedAt)},
					},
				},
			},
			{
				Put: &types.Put{
					TableName: aws.String(repo.tableName),
					Item: map[string]types.AttributeValue{
						"PK":     &types.AttributeValueMemberS{Value: "EXPENSE"},
						"SK":     &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", expense.ID)},
						"UserID": &types.AttributeValueMemberS{Value: expense.UserID},
					},
				},
			},
		},
	})
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "Error saving expense to dynamodb", err)
	}
	return expense.ID, nil
}

func (repo *ExpenseRepository) UpdateExpense(ctx context.Context, expense domain.Expense) error {
	expr, err := repo.buildUpdateExpression(expense)
	if err != nil {
		return err
	}
	_, err = repo.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "EXPENSE"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#%s", expense.UserID, expense.ID)},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "Error updating expense in dynamodb", err)
	}
	return nil
}

func (repo *ExpenseRepository) FindExpenseById(ctx context.Context, expenseId string) (domain.Expense, error) {
	userId, err := repo.findUserIdByExpenseId(ctx, expenseId)
	if err != nil {
		return domain.Expense{}, err
	}
	result, err := repo.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "EXPENSE"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#%s", userId, expenseId)},
		},
	})
	if err != nil {
		return domain.Expense{}, apperr.NewAppError(apperr.ErrInternal, "Error fetching expense from dynamodb", err)
	}
	if result.Item == nil {
		return domain.Expense{}, apperr.NewAppError(apperr.ErrNotFound, "Expense not found", nil)
	}
	expense, err := repo.toDomainExpense(result.Item)
	if err != nil {
		return domain.Expense{}, err
	}
	return expense, nil
}

func (repo *ExpenseRepository) FindAllExpenses(ctx context.Context, filterOptions domain.ExpensesFilterOptions) ([]domain.Expense, int, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "EXPENSE"},
			"SK": &types.AttributeValueMemberS{Value: "DETAILS#"},
		},
	})
	if err != nil {
		return []domain.Expense{}, 0, apperr.NewAppError(apperr.ErrInternal, "Error fetching expense from dynamodb", err)
	}
	expenses := []domain.Expense{}
	for _, item := range result.Items {
		expense, err := repo.toDomainExpense(item)
		if err != nil {
			return []domain.Expense{}, 0, err
		}
		expenses = append(expenses, expense)
	}
	return expenses, len(expenses), nil
}

func (repo *ExpenseRepository) GetExpenseSumByStatus(ctx context.Context, userID string, status domain.RequestStatus) (float64, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		FilterExpression:       aws.String("Status = :status"),
		ProjectionExpression:   aws.String("Amount"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "EXPENSE"},
			":sk":     &types.AttributeValueMemberS{Value: "DETAILS#"},
			":status": &types.AttributeValueMemberS{Value: string(status)},
		},
	})
	if err != nil {
		return 0, apperr.NewAppError(apperr.ErrInternal, "Error fetching expense from dynamodb", err)
	}
	expenseSum := 0.0
	for _, item := range result.Items {
		amountAttr, ok := item["Amount"].(*types.AttributeValueMemberN)
		if !ok {
			return 0, apperr.NewAppError(apperr.ErrInternal, "Error parsing Amount attribute", nil)
		}
		amount, err := strconv.ParseFloat(amountAttr.Value, 64)
		if err != nil {
			return 0, apperr.NewAppError(apperr.ErrInternal, "Error converting Amount to float", err)
		}
		expenseSum += amount
	}
	return expenseSum, nil
}

func (repo *ExpenseRepository) getBillsAttributeValue(bills []domain.Bill) types.AttributeValue {
	billsAttr := types.AttributeValueMemberL{
		Value: []types.AttributeValue{},
	}
	for _, bill := range bills {
		billID := bill.ID
		if billID == "" {
			billID = uuid.New().String()
		}
		billMap := types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"BillID":        &types.AttributeValueMemberS{Value: billID},
				"Amount":        &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", bill.Amount)},
				"Description":   &types.AttributeValueMemberS{Value: bill.Description},
				"AttachmentURL": &types.AttributeValueMemberS{Value: bill.AttachmentURL},
			},
		}
		billsAttr.Value = append(billsAttr.Value, &billMap)
	}
	return &billsAttr
}

type UpdateExpenseBill struct {
	BillID        string
	Amount        float64
	Description   string
	AttachmentURL string
}

func (repo *ExpenseRepository) buildUpdateExpression(expense domain.Expense) (expression.Expression, error) {
	bills := []UpdateExpenseBill{}
	for _, bill := range expense.Bills {
		bills = append(bills, UpdateExpenseBill{
			BillID:        bill.ID,
			Amount:        bill.Amount,
			Description:   bill.Description,
			AttachmentURL: bill.AttachmentURL,
		})
	}
	update := expression.Set(expression.Name("Amount"), expression.Value(expense.Amount))
	update.Set(expression.Name("Description"), expression.Value(expense.Description))
	update.Set(expression.Name("Purpose"), expression.Value(expense.Purpose))
	update.Set(expression.Name("IsReconciled"), expression.Value(expense.IsReconciled))
	update.Set(expression.Name("Status"), expression.Value(string(expense.Status)))
	update.Set(expression.Name("UpdatedAt"), expression.Value(time.Now().UnixMilli()))
	update.Set(expression.Name("ApprovedBy"), expression.Value(expense.ApprovedBy))
	update.Set(expression.Name("ApprovedAt"), expression.Value(expense.ApprovedAt))
	update.Set(expression.Name("ReviewedBy"), expression.Value(expense.ReviewedBy))
	update.Set(expression.Name("ReviewedAt"), expression.Value(expense.ReviewedAt))
	update.Set(expression.Name("Bills"), expression.Value(bills))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return expression.Expression{}, apperr.NewAppError(apperr.ErrInternal, "Error while building update expression", err)
	}
	return expr, nil
}

func (repo *ExpenseRepository) findUserIdByExpenseId(ctx context.Context, expenseId string) (string, error) {
	result, err := repo.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "EXPENSE"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", expenseId)},
		},
	})
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "Error fetching expense user mapping from dynamodb", err)
	}
	if result.Item == nil {
		return "", apperr.NewAppError(apperr.ErrNotFound, "Expense not found", nil)
	}
	userIDAttr, ok := result.Item["UserID"].(*types.AttributeValueMemberS)
	if !ok {
		return "", apperr.NewAppError(apperr.ErrInternal, "Error parsing UserID attribute", nil)
	}
	return userIDAttr.Value, nil
}

func (repo *ExpenseRepository) toDomainExpense(expenseItem map[string]types.AttributeValue) (domain.Expense, error) {
	createdAt, _ := strconv.ParseInt(expenseItem["CreatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	updatedAt, _ := strconv.ParseInt(expenseItem["UpdatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	amount, _ := strconv.ParseFloat(expenseItem["Amount"].(*types.AttributeValueMemberN).Value, 64)
	approvedAt, reviewedAt := 0, 0
	approvedBy, reviewedBy := "", ""
	if val, ok := expenseItem["ApprovedAt"].(*types.AttributeValueMemberN); ok {
		approvedAt, _ = strconv.Atoi(val.Value)
	}
	if val, ok := expenseItem["ReviewedAt"].(*types.AttributeValueMemberN); ok {
		reviewedAt, _ = strconv.Atoi(val.Value)
	}
	if val, ok := expenseItem["ApprovedBy"].(*types.AttributeValueMemberS); ok {
		approvedBy = val.Value
	}
	if val, ok := expenseItem["ReviewedBy"].(*types.AttributeValueMemberS); ok {
		reviewedBy = val.Value
	}

	bills := []domain.Bill{}
	if billsAttr, ok := expenseItem["Bills"].(*types.AttributeValueMemberL); ok {
		for _, billAttr := range billsAttr.Value {
			billMap := billAttr.(*types.AttributeValueMemberM).Value
			amount, _ := strconv.ParseFloat(billMap["Amount"].(*types.AttributeValueMemberN).Value, 64)
			bill := domain.Bill{
				ID:            billMap["BillID"].(*types.AttributeValueMemberS).Value,
				ExpenseID:     expenseItem["ExpenseID"].(*types.AttributeValueMemberS).Value,
				Amount:        amount,
				Description:   billMap["Description"].(*types.AttributeValueMemberS).Value,
				AttachmentURL: billMap["AttachmentURL"].(*types.AttributeValueMemberS).Value,
			}
			bills = append(bills, bill)
		}
	}

	expense := domain.Expense{
		ID:           expenseItem["ExpenseID"].(*types.AttributeValueMemberS).Value,
		UserID:       expenseItem["UserID"].(*types.AttributeValueMemberS).Value,
		Amount:       amount,
		Description:  expenseItem["Description"].(*types.AttributeValueMemberS).Value,
		Status:       domain.RequestStatus(expenseItem["Status"].(*types.AttributeValueMemberS).Value),
		Purpose:      expenseItem["Purpose"].(*types.AttributeValueMemberS).Value,
		ApprovedBy:   approvedBy,
		ApprovedAt:   int64(approvedAt),
		ReviewedBy:   reviewedBy,
		ReviewedAt:   int64(reviewedAt),
		IsReconciled: expenseItem["IsReconciled"].(*types.AttributeValueMemberBOOL).Value,
		Bills:        bills,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	return expense, nil
}
