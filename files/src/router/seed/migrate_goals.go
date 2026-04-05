package seed

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
	log "github.com/sirupsen/logrus"
)

type legacyGoal struct {
	Id       string `dynamodbav:"id"`
	SeasonId string `dynamodbav:"seasonId"`
	TeamId   string `dynamodbav:"teamId"`
}

func MigrateGoals(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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
			log.WithError(err).Error("scan goals failed")
			return utils.ErrorResponse(500, "scan failed", err)
		}

		for _, item := range result.Items {
			var g legacyGoal
			if err := attributevalue.UnmarshalMap(item, &g); err != nil {
				log.WithError(err).Warn("unmarshal goal skipped")
				skipped++
				continue
			}
			if g.SeasonId == "" {
				skipped++
				continue
			}
			if g.TeamId != "" {
				skipped++
				continue
			}

			teamId, err := db.GetTeamIdBySeasonId(ctx, g.SeasonId)
			if err != nil || teamId == "" {
				log.WithError(err).WithField("goalId", g.Id).Warn("could not resolve teamId")
				skipped++
				continue
			}

			if _, err := db.TagGoalToSeason(ctx, g.Id, g.SeasonId); err != nil {
				log.WithError(err).WithField("goalId", g.Id).Warn("tag goal failed")
				skipped++
				continue
			}

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
				log.WithError(err).WithField("goalId", g.Id).Warn("update goal failed")
				skipped++
				continue
			}

			migrated++
			log.WithField("goalId", g.Id).Info("migrated goal")
		}

		if result.LastEvaluatedKey == nil || len(result.LastEvaluatedKey) == 0 {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	log.WithFields(log.Fields{"migrated": migrated, "skipped": skipped}).Info("Migration complete")
	return utils.SuccessResponse(200, "Migration complete", map[string]int{"migrated": migrated, "skipped": skipped})
}

