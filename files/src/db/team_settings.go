package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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
