package db

import (
	"errors"
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
