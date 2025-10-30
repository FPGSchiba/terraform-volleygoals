package db

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/fpgschiba/volleygoals/models"
	"golang.org/x/net/context"
)

func CreateTeam(ctx context.Context, name string) (*models.Team, error) {
	client = GetClient()
	team := &models.Team{
		Name:      name,
		Status:    models.TeamStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	result, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(teamsTableName),
		Item:      team.ToAttributeValues(),
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Created team: %v", result)
	return team, nil
}
