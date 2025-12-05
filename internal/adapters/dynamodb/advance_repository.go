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

type AdvanceRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewAdvanceRepository(client *dynamodb.Client, tableName string) ports.AdvanceRepository {
	return &AdvanceRepository{
		client:    client,
		tableName: tableName,
	}
}

func (repo *AdvanceRepository) SaveAdvance(ctx context.Context, advance domain.Advance) (string, error) {
	if advance.ID == "" {
		advance.ID = uuid.New().String()
	}
	if advance.CreatedAt <= 0 {
		advance.CreatedAt = time.Now().UnixMilli()
		advance.UpdatedAt = advance.CreatedAt
	}
	_, err := repo.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName: aws.String(repo.tableName),
					Item: map[string]types.AttributeValue{
						"PK":                  &types.AttributeValueMemberS{Value: "ADVANCE"},
						"SK":                  &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#%s", advance.UserID, advance.ID)},
						"AdvanceID":           &types.AttributeValueMemberS{Value: advance.ID},
						"UserID":              &types.AttributeValueMemberS{Value: advance.UserID},
						"Amount":              &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", advance.Amount)},
						"Purpose":             &types.AttributeValueMemberS{Value: advance.Purpose},
						"Description":         &types.AttributeValueMemberS{Value: string(advance.Description)},
						"Status":              &types.AttributeValueMemberS{Value: string(advance.Status)},
						"ReconciledExpenseID": &types.AttributeValueMemberS{Value: advance.ReconciledExpenseID},
						"ApprovedBy":          &types.AttributeValueMemberS{Value: advance.ApprovedBy},
						"ApprovedAt":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", advance.ApprovedAt)},
						"ReviewedBy":          &types.AttributeValueMemberS{Value: advance.ReviewedBy},
						"ReviewedAt":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", advance.ReviewedAt)},
						"CreatedAt":           &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", advance.CreatedAt)},
						"UpdatedAt":           &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", advance.UpdatedAt)},
					},
				},
			},
			{
				Put: &types.Put{
					TableName: aws.String(repo.tableName),
					Item: map[string]types.AttributeValue{
						"PK":     &types.AttributeValueMemberS{Value: "ADVANCE"},
						"SK":     &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", advance.ID)},
						"UserID": &types.AttributeValueMemberS{Value: advance.UserID},
					},
				},
			},
		},
	})
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "Error saving advance to dynamodb", err)
	}
	return advance.ID, nil
}

func (repo *AdvanceRepository) UpdateAdvance(ctx context.Context, advance domain.Advance) error {
	expr, err := repo.buildUpdateExpression(advance)
	if err != nil {
		return err
	}
	_, err = repo.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "ADVANCE"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#%s", advance.UserID, advance.ID)},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return apperr.NewAppError(apperr.ErrInternal, "Error updating advance in dynamodb", err)
	}
	return nil
}

func (repo *AdvanceRepository) FindAdvanceById(ctx context.Context, advanceId string) (domain.Advance, error) {
	userID, err := repo.findUserIdByAdvanceId(ctx, advanceId)
	if err != nil {
		return domain.Advance{}, err
	}
	result, err := repo.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "ADVANCE"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#%s", userID, advanceId)},
		},
	})
	if err != nil {
		return domain.Advance{}, apperr.NewAppError(apperr.ErrInternal, "Error fetching advance from dynamodb", err)
	}
	if result.Item == nil {
		return domain.Advance{}, apperr.NewAppError(apperr.ErrNotFound, "Advance not found", nil)
	}
	advance, err := repo.toDomainAdvance(result.Item)
	if err != nil {
		return domain.Advance{}, err
	}
	return advance, nil
}

func (repo *AdvanceRepository) FindAllAdvances(ctx context.Context, filterOptions domain.AdvancesFilterOptions) ([]domain.Advance, int, error) {
	var err error
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "ADVANCE"},
			":sk": &types.AttributeValueMemberS{Value: "DETAILS#"},
		},
	}

	if filterOptions.UserID != "" {
		queryInput.ExpressionAttributeValues[":sk"] = &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s#", filterOptions.UserID)}
	}

	if filterOptions.Status != "" {
		queryInput.FilterExpression = aws.String("#status = :status")
		queryInput.ExpressionAttributeValues[":status"] = &types.AttributeValueMemberS{Value: string(filterOptions.Status)}
		queryInput.ExpressionAttributeNames = map[string]string{
			"#status": "Status",
		}
	}

	// total records
	queryInput.Select = types.SelectCount
	countResult, err := repo.client.Query(ctx, queryInput)
	if err != nil {
		return []domain.Advance{}, 0, apperr.NewAppError(apperr.ErrInternal, "Error fetching advances from dynamodb", err)
	}
	totalRecords := countResult.Count

	// fast cursor
	queryInput, _, err = offsetQuery(
		ctx,
		repo.client,
		queryInput,
		filterOptions.Page-1,
		filterOptions.Limit,
	)
	if err != nil {
		return []domain.Advance{}, 0, err
	}
	// if no results for the page
	if queryInput == nil {
		return []domain.Advance{}, 0, nil
	}

	// advances
	queryInput.Select = types.SelectAllAttributes
	queryInput.Limit = aws.Int32(int32(filterOptions.Limit))
	result, err := repo.client.Query(ctx, queryInput)
	if err != nil {
		return []domain.Advance{}, 0, apperr.NewAppError(apperr.ErrInternal, "Error fetching advance from dynamodb", err)
	}
	advances := []domain.Advance{}
	for _, item := range result.Items {
		advance, err := repo.toDomainAdvance(item)
		if err != nil {
			return []domain.Advance{}, 0, err
		}
		advances = append(advances, advance)
	}
	return advances, int(totalRecords), nil
}

func (repo *AdvanceRepository) GetAdvanceSumByStatus(ctx context.Context, userID string, status domain.RequestStatus) (float64, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		FilterExpression:       aws.String("#status = :status"),
		ProjectionExpression:   aws.String("Amount"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "ADVANCE"},
			":sk":     &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s", userID)},
			":status": &types.AttributeValueMemberS{Value: string(status)},
		},
		ExpressionAttributeNames: map[string]string{
			"#status": "Status",
		},
	})
	if err != nil {
		return 0, apperr.NewAppError(apperr.ErrInternal, "Error fetching advance from dynamodb", err)
	}
	advanceSum := 0.0
	for _, item := range result.Items {
		amountAttr, ok := item["Amount"].(*types.AttributeValueMemberN)
		if !ok {
			return 0, apperr.NewAppError(apperr.ErrInternal, "Error parsing Amount attribute", nil)
		}
		amount, err := strconv.ParseFloat(amountAttr.Value, 64)
		if err != nil {
			return 0, apperr.NewAppError(apperr.ErrInternal, "Error converting Amount to float", err)
		}
		advanceSum += amount
	}
	return advanceSum, nil
}

func (repo *AdvanceRepository) GetReconciledAdvancesSum(ctx context.Context, userID string) (float64, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		FilterExpression:       aws.String("attribute_exists(ReconciledExpenseID) AND size(ReconciledExpenseID) > :empty"),
		ProjectionExpression:   aws.String("Amount"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":    &types.AttributeValueMemberS{Value: "ADVANCE"},
			":sk":    &types.AttributeValueMemberS{Value: fmt.Sprintf("DETAILS#%s", userID)},
			":empty": &types.AttributeValueMemberN{Value: "0"},
		},
	})
	if err != nil {
		return 0, apperr.NewAppError(apperr.ErrInternal, "Error fetching advance from dynamodb", err)
	}
	advanceSum := 0.0
	for _, item := range result.Items {
		amountAttr, ok := item["Amount"].(*types.AttributeValueMemberN)
		if !ok {
			return 0, apperr.NewAppError(apperr.ErrInternal, "Error parsing Amount attribute", nil)
		}
		amount, err := strconv.ParseFloat(amountAttr.Value, 64)
		if err != nil {
			return 0, apperr.NewAppError(apperr.ErrInternal, "Error converting Amount to float", err)
		}
		advanceSum += amount
	}
	return advanceSum, nil
}

func (repo *AdvanceRepository) buildUpdateExpression(advance domain.Advance) (expression.Expression, error) {
	update := expression.Set(expression.Name("Amount"), expression.Value(advance.Amount))
	update.Set(expression.Name("Description"), expression.Value(advance.Description))
	update.Set(expression.Name("Purpose"), expression.Value(advance.Purpose))
	update.Set(expression.Name("Status"), expression.Value(string(advance.Status)))
	update.Set(expression.Name("UpdatedAt"), expression.Value(time.Now().UnixMilli()))
	update.Set(expression.Name("ApprovedBy"), expression.Value(advance.ApprovedBy))
	update.Set(expression.Name("ApprovedAt"), expression.Value(advance.ApprovedAt))
	update.Set(expression.Name("ReviewedBy"), expression.Value(advance.ReviewedBy))
	update.Set(expression.Name("ReviewedAt"), expression.Value(advance.ReviewedAt))
	update.Set(expression.Name("ReconciledExpenseID"), expression.Value(advance.ReconciledExpenseID))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return expression.Expression{}, apperr.NewAppError(apperr.ErrInternal, "Error while building update expression", err)
	}
	return expr, nil
}

func (repo *AdvanceRepository) findUserIdByAdvanceId(ctx context.Context, advanceId string) (string, error) {
	result, err := repo.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "ADVANCE"},
			"SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("USER#%s", advanceId)},
		},
	})
	if err != nil {
		return "", apperr.NewAppError(apperr.ErrInternal, "Error fetching advance user mapping from dynamodb", err)
	}
	if result.Item == nil {
		return "", apperr.NewAppError(apperr.ErrNotFound, "Advance not found", nil)
	}
	userIDAttr, ok := result.Item["UserID"].(*types.AttributeValueMemberS)
	if !ok {
		return "", apperr.NewAppError(apperr.ErrInternal, "Error parsing UserID attribute", nil)
	}
	return userIDAttr.Value, nil
}

func (repo *AdvanceRepository) toDomainAdvance(advanceItem map[string]types.AttributeValue) (domain.Advance, error) {
	createdAt, _ := strconv.ParseInt(advanceItem["CreatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	updatedAt, _ := strconv.ParseInt(advanceItem["UpdatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)
	amount, _ := strconv.ParseFloat(advanceItem["Amount"].(*types.AttributeValueMemberN).Value, 64)
	approvedAt, reviewedAt := 0, 0
	approvedBy, reviewedBy := "", ""
	reconciledExpenseID := ""
	if val, ok := advanceItem["ApprovedAt"].(*types.AttributeValueMemberN); ok {
		approvedAt, _ = strconv.Atoi(val.Value)
	}
	if val, ok := advanceItem["ReviewedAt"].(*types.AttributeValueMemberN); ok {
		reviewedAt, _ = strconv.Atoi(val.Value)
	}
	if val, ok := advanceItem["ApprovedBy"].(*types.AttributeValueMemberS); ok {
		approvedBy = val.Value
	}
	if val, ok := advanceItem["ReviewedBy"].(*types.AttributeValueMemberS); ok {
		reviewedBy = val.Value
	}
	if val, ok := advanceItem["ReconciledExpenseID"].(*types.AttributeValueMemberS); ok {
		reconciledExpenseID = val.Value
	}
	advance := domain.Advance{
		ID:                  advanceItem["AdvanceID"].(*types.AttributeValueMemberS).Value,
		UserID:              advanceItem["UserID"].(*types.AttributeValueMemberS).Value,
		Amount:              amount,
		Purpose:             advanceItem["Purpose"].(*types.AttributeValueMemberS).Value,
		Description:         advanceItem["Description"].(*types.AttributeValueMemberS).Value,
		Status:              domain.RequestStatus(advanceItem["Status"].(*types.AttributeValueMemberS).Value),
		ReconciledExpenseID: reconciledExpenseID,
		ApprovedBy:          approvedBy,
		ApprovedAt:          int64(approvedAt),
		ReviewedBy:          reviewedBy,
		ReviewedAt:          int64(reviewedAt),
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}
	return advance, nil
}
