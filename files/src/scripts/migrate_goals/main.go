//go:build ignore

// migrate_goals migrates all existing Goal records from the old SeasonId model
// to the new TeamId model, creating goal_seasons join records in the process.
//
// Run once per environment after deploying the schema changes:
//
//	go run -tags local files/src/scripts/migrate_goals/main.go
package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/db"
)

// legacyGoal mirrors the old Goal schema with SeasonId still present.
type legacyGoal struct {
	Id       string `dynamodbav:"id"`
	SeasonId string `dynamodbav:"seasonId"`
	TeamId   string `dynamodbav:"teamId"`
}

func main() {
	ctx := context.Background()
	db.InitClient(nil)
	client := db.GetClient()

	tableName := os.Getenv("GOALS_TABLE_NAME")
	if tableName == "" {
		tableName = "dev-goals"
	}

	migrated := 0
	skipped := 0

	var lastKey map[string]types.AttributeValue

	for {
		input := &dynamodb.ScanInput{
			TableName:            aws.String(tableName),
			ProjectionExpression: aws.String("id, seasonId, teamId"),
		}
		if lastKey != nil {
			input.ExclusiveStartKey = lastKey
		}

		result, err := client.Scan(ctx, input)
		if err != nil {
			log.Fatalf("scan goals: %v", err)
		}

		for _, item := range result.Items {
			var g legacyGoal
			if err := attributevalue.UnmarshalMap(item, &g); err != nil {
				log.Printf("unmarshal goal: %v — skipping", err)
				skipped++
				continue
			}
			if g.SeasonId == "" {
				skipped++
				continue
			}
			if g.TeamId != "" {
				// Already migrated
				skipped++
				continue
			}

			teamId, err := db.GetTeamIdBySeasonId(ctx, g.SeasonId)
			if err != nil || teamId == "" {
				log.Printf("could not resolve teamId for goal %s (seasonId=%s): %v — skipping", g.Id, g.SeasonId, err)
				skipped++
				continue
			}

			// 1. Write goal_seasons join record
			if _, err := db.TagGoalToSeason(ctx, g.Id, g.SeasonId); err != nil {
				log.Printf("tag goal %s to season %s: %v — skipping", g.Id, g.SeasonId, err)
				skipped++
				continue
			}

			// 2. Set teamId and clear seasonId on the Goal record
			_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName: aws.String(tableName),
				Key: map[string]types.AttributeValue{
					"id": item["id"],
				},
				UpdateExpression: aws.String("SET teamId = :tid REMOVE seasonId"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":tid": &types.AttributeValueMemberS{Value: teamId},
				},
			})
			if err != nil {
				log.Printf("update goal %s: %v — skipping", g.Id, err)
				skipped++
				continue
			}

			migrated++
			log.Printf("migrated goal %s (season %s → team %s)", g.Id, g.SeasonId, teamId)
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	log.Printf("Migration complete: %d migrated, %d skipped", migrated, skipped)
}
