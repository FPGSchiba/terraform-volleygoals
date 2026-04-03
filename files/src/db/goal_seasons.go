package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

func TagGoalToSeason(ctx context.Context, goalId, seasonId string) (*models.GoalSeason, error) {
	client = GetClient()
	gs := &models.GoalSeason{
		Id:        goalId + "#" + seasonId,
		GoalId:    goalId,
		SeasonId:  seasonId,
		CreatedAt: time.Now(),
	}
	item, err := attributevalue.MarshalMap(gs)
	if err != nil {
		return nil, fmt.Errorf("TagGoalToSeason: marshal: %w", err)
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &goalSeasonsTableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return nil, fmt.Errorf("TagGoalToSeason: goal %s already tagged to season %s", goalId, seasonId)
		}
		return nil, fmt.Errorf("TagGoalToSeason: put: %w", err)
	}
	return gs, nil
}

func UntagGoalFromSeason(ctx context.Context, goalId, seasonId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &goalSeasonsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: goalId + "#" + seasonId},
		},
	})
	if err != nil {
		return fmt.Errorf("UntagGoalFromSeason: %w", err)
	}
	return nil
}

func ListSeasonsByGoalId(ctx context.Context, goalId string) ([]*models.GoalSeason, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &goalSeasonsTableName,
		IndexName:              aws.String("goalIdIndex"),
		KeyConditionExpression: aws.String("goalId = :gid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gid": &types.AttributeValueMemberS{Value: goalId},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ListSeasonsByGoalId: %w", err)
	}
	var items []*models.GoalSeason
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, fmt.Errorf("ListSeasonsByGoalId: unmarshal: %w", err)
	}
	return items, nil
}

func ListGoalIdsBySeasonId(ctx context.Context, seasonId string) ([]string, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &goalSeasonsTableName,
		IndexName:              aws.String("seasonIdIndex"),
		KeyConditionExpression: aws.String("seasonId = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": &types.AttributeValueMemberS{Value: seasonId},
		},
		ProjectionExpression: aws.String("goalId"),
	})
	if err != nil {
		return nil, fmt.Errorf("ListGoalIdsBySeasonId: %w", err)
	}
	ids := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		if v, ok := item["goalId"].(*types.AttributeValueMemberS); ok {
			ids = append(ids, v.Value)
		}
	}
	return ids, nil
}
