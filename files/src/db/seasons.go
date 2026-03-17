package db

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

func CreateSeason(ctx context.Context, teamId, name string, start, end time.Time) (*models.Season, error) {
	client = GetClient()
	now := time.Now()
	var status models.SeasonStatus
	if start.Before(now) {
		status = models.SeasonStatusActive
	} else {
		status = models.SeasonStatusPlanned
	}

	season := &models.Season{
		Id:        models.GenerateID(),
		TeamId:    teamId,
		Name:      name,
		StartDate: start,
		EndDate:   end,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &seasonsTableName,
		Item:      season.ToAttributeValues(),
	})
	if err != nil {
		return nil, err
	}
	return season, nil
}

func GetSeasonById(ctx context.Context, seasonId string) (*models.Season, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &seasonsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: seasonId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var season models.Season
	err = attributevalue.UnmarshalMap(result.Item, &season)
	if err != nil {
		return nil, err
	}
	return &season, nil
}

func UpdateSeason(ctx context.Context, seasonId string, name *string, start, end *time.Time, status *models.SeasonStatus) (*models.Season, error) {
	client = GetClient()
	updateParts := make([]string, 0)
	exprAttrValues := make(map[string]types.AttributeValue)
	exprAttrNames := make(map[string]string)

	if name != nil {
		updateParts = append(updateParts, "#n = :name")
		exprAttrNames["#n"] = "name"
		exprAttrValues[":name"] = &types.AttributeValueMemberS{Value: *name}
	}
	if start != nil {
		updateParts = append(updateParts, "#sd = :startDate")
		exprAttrNames["#sd"] = "startDate"
		exprAttrValues[":startDate"] = &types.AttributeValueMemberS{Value: start.Format(time.RFC3339)}
	}
	if end != nil {
		updateParts = append(updateParts, "#ed = :endDate")
		exprAttrNames["#ed"] = "endDate"
		exprAttrValues[":endDate"] = &types.AttributeValueMemberS{Value: end.Format(time.RFC3339)}
	}
	if status != nil {
		updateParts = append(updateParts, "#s = :status")
		exprAttrNames["#s"] = "status"
		exprAttrValues[":status"] = &types.AttributeValueMemberS{Value: string(*status)}
	}

	// Always update UpdatedAt
	updateParts = append(updateParts, "#ua = :updatedAt")
	exprAttrNames["#ua"] = "updatedAt"
	exprAttrValues[":updatedAt"] = &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}

	updateExpr := "SET " + strings.Join(updateParts, ", ")

	result, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &seasonsTableName,
		Key:                       map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: seasonId}},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprAttrValues,
		ExpressionAttributeNames:  exprAttrNames,
		ReturnValues:              types.ReturnValueAllNew,
	})
	if err != nil {
		return nil, err
	}
	var updatedSeason models.Season
	err = attributevalue.UnmarshalMap(result.Attributes, &updatedSeason)
	if err != nil {
		return nil, err
	}
	return &updatedSeason, nil
}

func DeleteSeason(ctx context.Context, seasonId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &seasonsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: seasonId},
		},
	})
	return err
}

// ListSeasons returns a page of seasons according to SeasonFilter (limit, cursor, sorting, and optional filters).
func ListSeasons(ctx context.Context, filter SeasonFilter) ([]*models.Season, int, *models.Cursor, bool, error) {
	client = GetClient()

	// sane default limit
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}

	in := &dynamodb.ScanInput{
		TableName: aws.String(seasonsTableName),
		Limit:     aws.Int32(int32(limit)),
	}

	// Build filter expression using SeasonFilter
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

	seasons, err := unmarshalSeasons(result.Items)
	if err != nil {
		return nil, 0, nil, false, err
	}

	// Sorting (in-memory) using embedded FilterOptions
	if sortBy, sortOrder := filter.NormalizeSort(); sortBy != "" {
		sortSeasons(seasons, sortBy, sortOrder)
	}

	// Build cursor from LastEvaluatedKey
	nextCursor, hasMore := nextCursorFromLEK(result.LastEvaluatedKey)

	return seasons, len(seasons), nextCursor, hasMore, nil
}

func unmarshalSeasons(items []map[string]types.AttributeValue) ([]*models.Season, error) {
	seasons := make([]*models.Season, 0, len(items))
	for _, it := range items {
		var s models.Season
		if err := attributevalue.UnmarshalMap(it, &s); err != nil {
			return nil, err
		}
		seasons = append(seasons, &s)
	}
	return seasons, nil
}

func sortSeasons(seasons []*models.Season, sortBy, sortOrder string) {
	desc := strings.ToLower(sortOrder) == "desc"
	switch sortBy {
	case "name":
		sort.Slice(seasons, func(i, j int) bool {
			if desc {
				return seasons[i].Name > seasons[j].Name
			}
			return seasons[i].Name < seasons[j].Name
		})
	case "startdate", "startDate":
		sort.Slice(seasons, func(i, j int) bool {
			if desc {
				return seasons[i].StartDate.After(seasons[j].StartDate)
			}
			return seasons[i].StartDate.Before(seasons[j].StartDate)
		})
	case "enddate", "endDate":
		sort.Slice(seasons, func(i, j int) bool {
			if desc {
				return seasons[i].EndDate.After(seasons[j].EndDate)
			}
			return seasons[i].EndDate.Before(seasons[j].EndDate)
		})
	case "createdat", "createdAt":
		sort.Slice(seasons, func(i, j int) bool {
			if desc {
				return seasons[i].CreatedAt.After(seasons[j].CreatedAt)
			}
			return seasons[i].CreatedAt.Before(seasons[j].CreatedAt)
		})
	}
}

// GetAllSeasonIdsByTeamId returns the full set of season IDs belonging to a team.
// It paginates through all DynamoDB pages so the result is complete.
func GetAllSeasonIdsByTeamId(ctx context.Context, teamId string) (map[string]struct{}, error) {
	seasonIds := make(map[string]struct{})
	var cursor *models.Cursor
	for {
		seasons, _, nextCursor, hasMore, err := ListSeasons(ctx, SeasonFilter{
			FilterOptions: FilterOptions{Limit: 100, Cursor: cursor},
			TeamId:        teamId,
		})
		if err != nil {
			return nil, err
		}
		for _, s := range seasons {
			seasonIds[s.Id] = struct{}{}
		}
		if !hasMore {
			break
		}
		cursor = nextCursor
	}
	return seasonIds, nil
}

func GetTeamIdBySeasonId(ctx context.Context, seasonId string) (string, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &seasonsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: seasonId},
		},
		ProjectionExpression: aws.String("teamId"),
	})
	if err != nil {
		return "", err
	}
	if result.Item == nil {
		return "", nil
	}
	var season models.Season
	err = attributevalue.UnmarshalMap(result.Item, &season)
	if err != nil {
		return "", err
	}
	return season.TeamId, nil
}
