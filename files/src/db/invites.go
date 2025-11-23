package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

func CreateInvite(ctx context.Context, teamId, email, inviterSub, token string, role models.TeamMemberRole) (*models.Invite, error) {
	client = GetClient()
	invite := models.Invite{
		Id:        models.GenerateID(),
		TeamId:    teamId,
		Email:     email,
		Role:      role,
		Status:    models.InviteStatusPending,
		InvitedBy: inviterSub,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // Expires in 7 days
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(invitesTableName),
		Item:      invite.ToAttributeValues(),
	})
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func RemoveInviteById(ctx context.Context, inviteId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(invitesTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: inviteId},
		},
	})
	return err
}
