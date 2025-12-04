package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

func offsetQuery(ctx context.Context, client *dynamodb.Client, queryInput *dynamodb.QueryInput, page, limit int) (*dynamodb.QueryInput, map[string]types.AttributeValue, error) {
	if page <= 0 {
		return queryInput, nil, nil
	}

	prevSelect := queryInput.Select
	prevProjection := queryInput.ProjectionExpression

	queryInput.Select = types.SelectSpecificAttributes
	queryInput.ProjectionExpression = aws.String("PK, SK")

	offset := limit * page // 0-based page index
	var lastEvaluatedKey map[string]types.AttributeValue = nil
	for offset > 0 {
		queryInput.ExclusiveStartKey = lastEvaluatedKey
		queryInput.Limit = aws.Int32(int32(offset))
		result, err := client.Query(ctx, queryInput)
		if err != nil {
			return nil, nil, apperr.NewAppError(apperr.ErrInternal, "Error paginating from dynamodb", err)
		}
		lastEvaluatedKey = result.LastEvaluatedKey
		offset -= len(result.Items)
		if lastEvaluatedKey == nil {
			break
		}
	}

	if lastEvaluatedKey == nil {
		return nil, nil, nil
	}

	queryInput.Select = prevSelect
	queryInput.ProjectionExpression = prevProjection
	queryInput.ExclusiveStartKey = lastEvaluatedKey
	return queryInput, lastEvaluatedKey, nil
}
