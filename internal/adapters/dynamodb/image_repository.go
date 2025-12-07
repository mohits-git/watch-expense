package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mohits-git/watch-expense/internal/ports"
	"github.com/mohits-git/watch-expense/internal/utils/apperr"
)

type ImageMetadataRepository struct {
  client *dynamodb.Client
  tableName string
}

func NewImageMetadataRepository(client *dynamodb.Client, tableName string) ports.ImageMetadataRepository {
  return &ImageMetadataRepository{
    client: client,
    tableName: tableName,
  }
}

func (imr *ImageMetadataRepository) SaveImageUserMetadata(ctx context.Context, imageURL string, userID string) error {
  _, err := imr.client.PutItem(ctx, &dynamodb.PutItemInput{
    TableName: aws.String(imr.tableName),
    Item: map[string]types.AttributeValue{
      "PK": &types.AttributeValueMemberS{Value: "IMAGE"},
      "SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("IMAGE#%s", imageURL)},
      "ImageURL": &types.AttributeValueMemberS{Value: imageURL},
      "UserID":  &types.AttributeValueMemberS{Value: userID},
    },
  })
  if err != nil {
    return apperr.NewAppError(apperr.ErrInternal, "failed to save image metadata", err)
  }
  return nil
}

func (imr *ImageMetadataRepository) GetImageUserMetadata(ctx context.Context, imageURL string) (userID string, err error) {
  result, err := imr.client.GetItem(ctx, &dynamodb.GetItemInput{
    TableName: aws.String(imr.tableName),
    Key: map[string]types.AttributeValue{
      "PK": &types.AttributeValueMemberS{Value: "IMAGE"},
      "SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("IMAGE#%s", imageURL)},
    },
  })
  if err != nil {
    return "", apperr.NewAppError(apperr.ErrInternal, "failed to get image metadata", err)
  }
  if result.Item == nil {
    return "", apperr.NewAppError(apperr.ErrNotFound, "image metadata not found", nil)
  }

  userIDAttr, ok := result.Item["UserID"].(*types.AttributeValueMemberS)
  if !ok {
    return "", apperr.NewAppError(apperr.ErrInternal, "invalid userID attribute type", nil)
  }

  return userIDAttr.Value, nil
}

func (imr *ImageMetadataRepository) DeleteImageUserMetadata(ctx context.Context, imageURL string) error {
  _, err := imr.GetImageUserMetadata(ctx, imageURL)
  if err != nil {
    return err
  }
  _, err = imr.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
    TableName: aws.String(imr.tableName),
    Key: map[string]types.AttributeValue{
      "PK": &types.AttributeValueMemberS{Value: "IMAGE"},
      "SK": &types.AttributeValueMemberS{Value: fmt.Sprintf("IMAGE#%s", imageURL)},
    },
  })
  if err != nil {
    return apperr.NewAppError(apperr.ErrInternal, "failed to delete image metadata", err)
  }
  return nil
}
