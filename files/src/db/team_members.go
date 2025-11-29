package db

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
	log "github.com/sirupsen/logrus"
)

func GetTeamMemberByUserIDAndTeamID(ctx context.Context, userID string, teamID string) (*models.TeamMember, error) {
	client = GetClient()
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

func GetMembershipsByUserID(ctx context.Context, userID string) ([]*models.TeamMember, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName: &teamMembersTableName,
		IndexName: aws.String("userIdIndex"),
		KeyConditions: map[string]types.Condition{
			"userId": {
				ComparisonOperator: "EQ",
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: userID},
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
	})
	if err != nil {
		return nil, err
	}
	var teamMembers []*models.TeamMember
	err = attributevalue.UnmarshalListOfMaps(result.Items, &teamMembers)
	if err != nil {
		return nil, err
	}
	return teamMembers, nil
}

func DeleteTeamMembershipsByTeamID(ctx context.Context, teamID string) error {
	memberships, err := GetMembershipsByTeamID(ctx, teamID)
	if err != nil {
		return err
	}
	for _, membership := range memberships {
		_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &teamMembersTableName,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: membership.Id},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func DeleteTeamMembershipsByUserID(ctx context.Context, userID string) error {
	memberships, err := GetMembershipsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, membership := range memberships {
		_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &teamMembersTableName,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: membership.Id},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func GetMembershipsByTeamID(ctx context.Context, teamID string) ([]*models.TeamMember, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName: &teamMembersTableName,
		IndexName: aws.String("teamIdIndex"),
		KeyConditions: map[string]types.Condition{
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
	})
	if err != nil {
		return nil, err
	}
	var teamMembers []*models.TeamMember
	err = attributevalue.UnmarshalListOfMaps(result.Items, &teamMembers)
	if err != nil {
		return nil, err
	}
	return teamMembers, nil
}

func CreateTeamMemberFromInvite(ctx context.Context, invite *models.Invite) (*models.TeamMember, error) {
	log.Printf("[DEBUG] CreateTeamMemberFromInvite AcceptedBy: %v", invite.AcceptedBy)
	client = GetClient()
	timeNow := time.Now()
	if invite.AcceptedBy == nil {
		return nil, fmt.Errorf("CreateTeamMemberFromInvite: invite.AcceptedBy is nil")
	}
	teamMember := &models.TeamMember{
		Id:        models.GenerateID(),
		TeamId:    invite.TeamId,
		UserId:    *invite.AcceptedBy,
		Role:      invite.Role,
		Status:    models.TeamMemberStatusActive,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		JoinedAt:  &timeNow,
	}
	item, err := attributevalue.MarshalMap(teamMember)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &teamMembersTableName,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	return teamMember, nil
}

func IsUserMemberOfTeam(ctx context.Context, userID string, teamID string) (bool, error) {
	teamMember, err := GetTeamMemberByUserIDAndTeamID(ctx, userID, teamID)
	if err != nil {
		return false, err
	}
	return teamMember != nil, nil
}

func GetTeamAssignmentsByUserID(ctx context.Context, userID string) ([]*models.TeamAssignment, error) {
	teamMembers, err := GetMembershipsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	var teamAssignments []*models.TeamAssignment
	for _, tm := range teamMembers {
		team, err := GetTeamById(ctx, tm.TeamId)
		if err != nil {
			return nil, err
		}
		if team == nil {
			log.Printf("[WARN] GetTeamAssignmentsByUserID: team not found for teamId %s", tm.TeamId)
			continue
		}
		teamAssignment := &models.TeamAssignment{
			Team:   *team,
			Role:   tm.Role,
			Status: tm.Status,
		}
		teamAssignments = append(teamAssignments, teamAssignment)
	}
	return teamAssignments, nil
}
