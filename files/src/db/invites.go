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

const InviteExpiresInDays = 7

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
		ExpiresAt: time.Now().Add(InviteExpiresInDays * 24 * time.Hour), // Expires in 7 days
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

func DoesInviteExistByToken(ctx context.Context, token string) (bool, error) {
	client = GetClient()

	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(invitesTableName),
		IndexName:              aws.String("tokenIndex"),
		KeyConditionExpression: aws.String("#inviteToken = :inviteToken"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inviteToken": &types.AttributeValueMemberS{Value: token},
			":status":      &types.AttributeValueMemberS{Value: string(models.InviteStatusPending)},
		},
		FilterExpression: aws.String("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status":      "status",
			"#inviteToken": "inviteToken",
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return false, err
	}
	return len(result.Items) > 0, nil
}

func GetInviteByToken(ctx context.Context, token string) (*models.Invite, error) {
	client = GetClient()

	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(invitesTableName),
		IndexName:              aws.String("tokenIndex"),
		KeyConditionExpression: aws.String("#inviteToken = :inviteToken"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":inviteToken": &types.AttributeValueMemberS{Value: token},
			":status":      &types.AttributeValueMemberS{Value: string(models.InviteStatusPending)},
		},
		FilterExpression: aws.String("#status = :status"),
		ExpressionAttributeNames: map[string]string{
			"#status":      "status",
			"#inviteToken": "inviteToken",
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var invite models.Invite
	err = attributevalue.UnmarshalMap(result.Items[0], &invite)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func CompleteInvite(ctx context.Context, inviteId, acceptedBy string, accept bool) (*models.Invite, error) {
	client = GetClient()
	updateExpr := "SET #status = :status, #updatedAt = :updatedAt, #acceptedBy = :acceptedBy, #acceptedAt = :acceptedAt"
	exprAttrValues := map[string]types.AttributeValue{
		":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}
	exprAttrNames := map[string]string{
		"#status":     "status",
		"#updatedAt":  "updatedAt",
		"#acceptedBy": "acceptedBy",
	}
	if accept {
		exprAttrValues[":status"] = &types.AttributeValueMemberS{Value: string(models.InviteStatusAccepted)}
		exprAttrValues[":acceptedAt"] = &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}
		exprAttrValues[":acceptedBy"] = &types.AttributeValueMemberS{Value: acceptedBy}
		exprAttrNames["#acceptedAt"] = "acceptedAt"
	} else {
		updateExpr = "SET #status = :status, #updatedAt = :updatedAt, #acceptedBy = :acceptedBy, #declinedAt = :declinedAt"
		exprAttrValues[":status"] = &types.AttributeValueMemberS{Value: string(models.InviteStatusDeclined)}
		exprAttrValues[":declinedAt"] = &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}
		exprAttrNames["#declinedAt"] = "declinedAt"
	}

	response, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(invitesTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: inviteId},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprAttrValues,
		ExpressionAttributeNames:  exprAttrNames,
	})
	if err != nil {
		return nil, err
	}
	var updatedInvite models.Invite
	err = attributevalue.UnmarshalMap(response.Attributes, &updatedInvite)
	if err != nil {
		return nil, err
	}
	return &updatedInvite, nil
}
