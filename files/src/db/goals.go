package db

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

func CreateGoal(ctx context.Context, teamId string, ownerId string, goalType models.GoalType, title string, description string) (*models.Goal, error) {
	client = GetClient()
	now := time.Now()
	goal := &models.Goal{
		Id:          models.GenerateID(),
		TeamId:      teamId,
		OwnerId:     ownerId,
		GoalType:    goalType,
		Title:       title,
		Description: description,
		Status:      models.GoalStatusOpen,
		CreatedBy:   ownerId,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &goalsTableName,
		Item:      goal.ToAttributeValues(),
	})
	if err != nil {
		return nil, err
	}

	return goal, nil
}

func GetGoalById(ctx context.Context, goalId string) (*models.Goal, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &goalsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: goalId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var goal models.Goal
	err = attributevalue.UnmarshalMap(result.Item, &goal)
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

func UpdateGoal(ctx context.Context, goalId string, ownerId *string, title *string, description *string, status *models.GoalStatus) (*models.Goal, error) {
	client = GetClient()
	updateExpr := "SET updatedAt = :updatedAt"
	exprAttrValues := map[string]types.AttributeValue{
		":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}
	exprAttrNames := map[string]string{}

	if ownerId != nil {
		updateExpr += ", ownerId = :ownerId"
		exprAttrValues[":ownerId"] = &types.AttributeValueMemberS{Value: *ownerId}
	}

	if title != nil {
		updateExpr += ", title = :title"
		exprAttrValues[":title"] = &types.AttributeValueMemberS{Value: *title}
	}

	if description != nil {
		updateExpr += ", description = :description"
		exprAttrValues[":description"] = &types.AttributeValueMemberS{Value: *description}
	}

	if status != nil {
		updateExpr += ", #st = :status"
		exprAttrValues[":status"] = &types.AttributeValueMemberS{Value: string(*status)}
		exprAttrNames["#st"] = "status"
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 &goalsTableName,
		Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: goalId}},
		UpdateExpression:          &updateExpr,
		ExpressionAttributeValues: exprAttrValues,
		ReturnValues:              types.ReturnValueAllNew,
	}
	if len(exprAttrNames) > 0 {
		input.ExpressionAttributeNames = exprAttrNames
	}

	result, err := client.UpdateItem(ctx, input)

	if err != nil {
		return nil, err
	}

	var updatedGoal models.Goal
	err = attributevalue.UnmarshalMap(result.Attributes, &updatedGoal)
	if err != nil {
		return nil, err
	}

	return &updatedGoal, nil
}

func DeleteGoal(ctx context.Context, goalId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &goalsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: goalId},
		},
	})
	return err
}

func UpdateGoalPicture(ctx context.Context, goalId string, pictureUrl string) error {
	client = GetClient()
	_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &goalsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: goalId},
		},
		UpdateExpression: aws.String("SET picture = :picture, updatedAt = :updatedAt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":picture":   &types.AttributeValueMemberS{Value: pictureUrl},
			":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	return err
}

// ListGoals returns a page of goals according to GoalFilter (limit, cursor, sorting, and optional filters).
func ListGoals(ctx context.Context, filter GoalFilter) ([]*models.Goal, int, *models.Cursor, bool, error) {
	client = GetClient()

	// sane default limit
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}

	in := &dynamodb.ScanInput{
		TableName: aws.String(goalsTableName),
		Limit:     aws.Int32(int32(limit)),
	}

	// Build filter expression using GoalFilter
	expr, vals, names := filter.BuildExpression()
	if strings.TrimSpace(expr) != "" {
		in.FilterExpression = aws.String(expr)
		in.ExpressionAttributeValues = vals
		if len(names) > 0 {
			in.ExpressionAttributeNames = names
		}
	}

	// Resume from cursor if provided
	if filter.Cursor != nil && filter.Cursor.LastID != "" {
		in.ExclusiveStartKey = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: filter.Cursor.LastID},
		}
	}

	result, err := client.Scan(ctx, in)
	if err != nil {
		return nil, 0, nil, false, err
	}

	goals, err := unmarshalGoals(result.Items)
	if err != nil {
		return nil, 0, nil, false, err
	}

	// Sorting (in-memory) using embedded FilterOptions
	if sortBy, sortOrder := filter.NormalizeSort(); sortBy != "" {
		sortGoals(goals, sortBy, sortOrder)
	}

	// Build cursor from LastEvaluatedKey
	nextCursor, hasMore := nextCursorFromLEK(result.LastEvaluatedKey)

	return goals, len(goals), nextCursor, hasMore, nil
}

func CountGoalsBySeasonId(ctx context.Context, seasonId string) (total int, completed int, open int, inProgress int, err error) {
	goalIds, err := ListGoalIdsBySeasonId(ctx, seasonId)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("CountGoalsBySeasonId: list goal IDs: %w", err)
	}
	for _, gid := range goalIds {
		goal, gerr := GetGoalById(ctx, gid)
		if gerr != nil {
			return 0, 0, 0, 0, fmt.Errorf("CountGoalsBySeasonId: get goal %s: %w", gid, gerr)
		}
		if goal == nil || goal.Status == models.GoalStatusArchived {
			continue
		}
		total++
		switch goal.Status {
		case models.GoalStatusCompleted:
			completed++
		case models.GoalStatusOpen:
			open++
		case models.GoalStatusInProgress:
			inProgress++
		}
	}
	return total, completed, open, inProgress, nil
}

// SearchGoalsForTeam scans all goals whose title contains query (case-insensitive)
// and whose teamId matches the given team. Archived goals are excluded.
// Returns at most limit results.
func SearchGoalsForTeam(ctx context.Context, teamId, query string, limit int) ([]*models.Goal, error) {
	client = GetClient()
	queryLower := strings.ToLower(query)
	results := make([]*models.Goal, 0, limit)
	var lastKey map[string]types.AttributeValue

	for {
		in := &dynamodb.ScanInput{
			TableName:        aws.String(goalsTableName),
			FilterExpression: aws.String("teamId = :tid AND #st <> :archived"),
			ExpressionAttributeNames: map[string]string{
				"#st": "status",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":tid":      &types.AttributeValueMemberS{Value: teamId},
				":archived": &types.AttributeValueMemberS{Value: string(models.GoalStatusArchived)},
			},
		}
		if lastKey != nil {
			in.ExclusiveStartKey = lastKey
		}
		result, scanErr := client.Scan(ctx, in)
		if scanErr != nil {
			return nil, fmt.Errorf("SearchGoalsForTeam: scan: %w", scanErr)
		}
		for _, item := range result.Items {
			if len(results) >= limit {
				break
			}
			titleV, ok := item["title"].(*types.AttributeValueMemberS)
			if !ok {
				continue
			}
			if !strings.Contains(strings.ToLower(titleV.Value), queryLower) {
				continue
			}
			var g models.Goal
			if err := attributevalue.UnmarshalMap(item, &g); err != nil {
				return nil, fmt.Errorf("SearchGoalsForTeam: unmarshal: %w", err)
			}
			results = append(results, &g)
		}
		if len(results) >= limit || result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}
	return results, nil
}

func unmarshalGoals(items []map[string]types.AttributeValue) ([]*models.Goal, error) {
	goals := make([]*models.Goal, 0, len(items))
	for _, it := range items {
		var g models.Goal
		if err := attributevalue.UnmarshalMap(it, &g); err != nil {
			return nil, err
		}
		goals = append(goals, &g)
	}
	return goals, nil
}

func sortGoals(goals []*models.Goal, sortBy, sortOrder string) {
	desc := strings.ToLower(sortOrder) == "desc"
	switch sortBy {
	case "title":
		sort.Slice(goals, func(i, j int) bool {
			if desc {
				return goals[i].Title > goals[j].Title
			}
			return goals[i].Title < goals[j].Title
		})
	case "createdat", "createdAt":
		sort.Slice(goals, func(i, j int) bool {
			if desc {
				return goals[i].CreatedAt.After(goals[j].CreatedAt)
			}
			return goals[i].CreatedAt.Before(goals[j].CreatedAt)
		})
	case "updatedat", "updatedAt":
		sort.Slice(goals, func(i, j int) bool {
			if desc {
				return goals[i].UpdatedAt.After(goals[j].UpdatedAt)
			}
			return goals[i].UpdatedAt.Before(goals[j].UpdatedAt)
		})
	}
}
