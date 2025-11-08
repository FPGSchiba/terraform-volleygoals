package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

func createTeamSettings(ctx context.Context, teamId string) error {
	client = GetClient()
	teamSettings := &models.TeamSettings{
		Id:                          models.GenerateID(),
		TeamID:                      teamId,
		AllowFileUploads:            true,
		AllowTeamGoalComments:       true,
		AllowIndividualGoalComments: true,
		CreatedAt:                   time.Now(),
		UpdatedAt:                   time.Now(),
	}
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(teamSettingsTableName),
		Item:      teamSettings.ToAttributeValues(),
	})
	if err != nil {
		return err
	}
	return nil
}

func GetTeamSettingsByTeamID(ctx context.Context, teamId string) (*models.TeamSettings, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(teamSettingsTableName),
		IndexName:              aws.String("teamIdIndex"),
		KeyConditionExpression: aws.String("teamId = :teamId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":teamId": &types.AttributeValueMemberS{Value: teamId},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var teamSettings models.TeamSettings
	err = attributevalue.UnmarshalMap(result.Items[0], &teamSettings)
	if err != nil {
		return nil, err
	}
	return &teamSettings, nil
}
