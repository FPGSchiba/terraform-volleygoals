package db

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
	"golang.org/x/net/context"
)

func findTeamByName(ctx context.Context, name string) (*models.Team, error) {
	client = GetClient()
	result, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(teamsTableName),
		FilterExpression: aws.String("teamName = :name"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name": &types.AttributeValueMemberS{Value: name},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var team models.Team
	err = attributevalue.UnmarshalMap(result.Items[0], &team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// ListTeams returns a page of teams according to TeamFilter (which now contains Limit and Cursor).
func ListTeams(ctx context.Context, filter TeamFilter) ([]*models.Team, int, *models.Cursor, bool, error) {
	client = GetClient()

	// ensure sane default limit
	limit := filter.Limit
	if limit <= 0 {
		limit = 25
	}

	in := &dynamodb.ScanInput{
		TableName: aws.String(teamsTableName),
		Limit:     aws.Int32(int32(limit)),
	}

	// Build filter expression using TeamFilter
	if expr, vals, names := filter.BuildExpression(); expr != "" {
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

	teams, err := unmarshalTeams(result.Items)
	if err != nil {
		return nil, 0, nil, false, err
	}

	// Sorting (in-memory) using embedded FilterOptions
	if sortBy, sortOrder := filter.NormalizeSort(); sortBy != "" {
		sortTeams(teams, sortBy, sortOrder)
	}

	// Build cursor from LastEvaluatedKey
	nextCursor, hasMore := nextCursorFromLEK(result.LastEvaluatedKey)

	return teams, len(teams), nextCursor, hasMore, nil
}

func unmarshalTeams(items []map[string]types.AttributeValue) ([]*models.Team, error) {
	teams := make([]*models.Team, 0, len(items))
	for _, it := range items {
		var t models.Team
		if err := attributevalue.UnmarshalMap(it, &t); err != nil {
			return nil, err
		}
		teams = append(teams, &t)
	}
	return teams, nil
}

func sortTeams(teams []*models.Team, sortBy, sortOrder string) {
	desc := strings.ToLower(sortOrder) == "desc"
	switch sortBy {
	case "name":
		sort.Slice(teams, func(i, j int) bool {
			if desc {
				return teams[i].Name > teams[j].Name
			}
			return teams[i].Name < teams[j].Name
		})
	case "createdat", "createdAt":
		sort.Slice(teams, func(i, j int) bool {
			if desc {
				return teams[i].CreatedAt.After(teams[j].CreatedAt)
			}
			return teams[i].CreatedAt.Before(teams[j].CreatedAt)
		})
	}
}

func nextCursorFromLEK(lek map[string]types.AttributeValue) (*models.Cursor, bool) {
	if lek == nil || len(lek) == 0 {
		return nil, false
	}
	c := &models.Cursor{}
	if v, ok := lek["id"]; ok {
		if s, ok := v.(*types.AttributeValueMemberS); ok {
			c.LastID = s.Value
		}
	}
	if v, ok := lek["createdAt"]; ok {
		if s, ok := v.(*types.AttributeValueMemberS); ok {
			c.LastCreatedAt = s.Value
		}
	}
	return c, true
}

func CreateTeam(ctx context.Context, name string) (*models.Team, error) {
	client = GetClient()
	existingTeam, err := findTeamByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if existingTeam != nil {
		return nil, errors.New("team already exists")
	}
	team := &models.Team{
		Id:        models.GenerateID(),
		Name:      name,
		Status:    models.TeamStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = createTeamSettings(ctx, team.Id)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(teamsTableName),
		Item:      team.ToAttributeValues(),
	})
	if err != nil {
		return nil, err
	}
	return team, nil
}

func GetTeamById(ctx context.Context, teamId string) (*models.Team, error) {
	client = GetClient() // Now returns the instrumented client

	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(teamsTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: teamId},
		},
	})

	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}

	var team models.Team
	err = attributevalue.UnmarshalMap(result.Item, &team)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

func UpdateTeam(ctx context.Context, team *models.Team) error {
	client = GetClient()
	team.UpdatedAt = time.Now()
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(teamsTableName),
		Item:      team.ToAttributeValues(),
	})
	return err
}

func DeleteTeamByID(ctx context.Context, teamId string) error {
	// TODO: Delete related team settings, users, goals, etc.
	err := DeleteTeamSettingsByTeamID(ctx, teamId)
	if err != nil {
		return err
	}
	client = GetClient()
	_, err = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(teamsTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: teamId},
		},
	})
	return err
}
