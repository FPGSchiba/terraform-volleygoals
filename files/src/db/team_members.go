package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

func GetTeamMemberByUserIDAndTeamID(ctx context.Context, userID string, teamID string) (*models.TeamMember, error) {
	client := GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName: &teamMembersTableName,
		IndexName: aws.String("teamUserIdIndex"),
		KeyConditions: map[string]types.Condition{
			"userId": {
				ComparisonOperator: "EQ",
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: userID},
				},
			},
			"teamId": {
				ComparisonOperator: "EQ",
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: teamID},
				},
			},
		},
		FilterExpression: aws.String("#status = :activeStatus"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":activeStatus": &types.AttributeValueMemberS{Value: string(models.TeamMemberStatusActive)},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var teamMember models.TeamMember
	err = attributevalue.UnmarshalMap(result.Items[0], &teamMember)
	if err != nil {
		return nil, err
	}
	return &teamMember, nil
}

func HasRoleOnTeam(ctx context.Context, userID string, teamID string, role models.TeamMemberRole) (bool, error) {
	teamMember, err := GetTeamMemberByUserIDAndTeamID(ctx, userID, teamID)
	if err != nil {
		return false, err
	}
	if teamMember != nil && teamMember.Role == role {
		return true, nil
	}
	return false, nil
}
