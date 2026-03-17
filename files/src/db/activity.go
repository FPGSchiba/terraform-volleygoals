package db

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
	log "github.com/sirupsen/logrus"
)

func createActivity(ctx context.Context, activity *models.Activity) error {
	client = GetClient()
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &activitiesTableName,
		Item:      activity.ToAttributeValues(),
	})
	return err
}

// EmitActivity writes an activity record fire-and-forget. Errors are logged only.
func EmitActivity(ctx context.Context, activity *models.Activity) {
	go func() {
		if err := createActivity(ctx, activity); err != nil {
			log.WithError(err).Warn("failed to emit activity")
		}
	}()
}

func ListTeamActivities(ctx context.Context, filter ActivityFilter) ([]*models.Activity, int, *models.Cursor, bool, error) {
	client = GetClient()
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}

	in := &dynamodb.ScanInput{
		TableName: aws.String(activitiesTableName),
		Limit:     aws.Int32(int32(limit)),
	}

	expr, vals, names := filter.BuildExpression()
	if expr != "" {
		in.FilterExpression = aws.String(expr)
		in.ExpressionAttributeValues = vals
		if len(names) > 0 {
			in.ExpressionAttributeNames = names
		}
	}

	if filter.Cursor != nil && filter.Cursor.LastID != "" {
		in.ExclusiveStartKey = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: filter.Cursor.LastID},
		}
	}

	result, err := client.Scan(ctx, in)
	if err != nil {
		return nil, 0, nil, false, err
	}

	activities := make([]*models.Activity, 0, len(result.Items))
	for _, item := range result.Items {
		var a models.Activity
		if err := attributevalue.UnmarshalMap(item, &a); err != nil {
			return nil, 0, nil, false, err
		}
		activities = append(activities, &a)
	}

	// Sort by timestamp descending
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.After(activities[j].Timestamp)
	})

	nextCursor, hasMore := nextCursorFromLEK(result.LastEvaluatedKey)
	return activities, len(activities), nextCursor, hasMore, nil
}
