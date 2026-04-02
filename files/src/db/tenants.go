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
)

func CreateTenant(ctx context.Context, name, ownerId string) (*models.Tenant, error) {
	client = GetClient()
	now := time.Now()
	tenant := &models.Tenant{
		Id:        models.GenerateID(),
		Name:      name,
		OwnerId:   ownerId,
		CreatedAt: now,
		UpdatedAt: now,
	}
	item, err := attributevalue.MarshalMap(tenant)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tenantsTableName,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	if _, err = AddTenantMember(ctx, tenant.Id, ownerId, models.TenantMemberRoleAdmin); err != nil {
		return nil, fmt.Errorf("CreateTenant: add owner as admin: %w", err)
	}
	return tenant, nil
}

func GetTenantById(ctx context.Context, tenantId string) (*models.Tenant, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tenantsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: tenantId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var tenant models.Tenant
	err = attributevalue.UnmarshalMap(result.Item, &tenant)
	return &tenant, err
}

func DeleteTenantById(ctx context.Context, tenantId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &tenantsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: tenantId},
		},
	})
	return err
}

func AddTenantMember(ctx context.Context, tenantId, userId string, role models.TenantMemberRole) (*models.TenantMember, error) {
	client = GetClient()
	now := time.Now()
	member := &models.TenantMember{
		Id:        tenantId + "#" + userId,
		TenantId:  tenantId,
		UserId:    userId,
		Role:      role,
		Status:    models.TenantMemberStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	item, err := attributevalue.MarshalMap(member)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tenantMembersTableName,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	return member, nil
}

func RemoveTenantMember(ctx context.Context, memberId string) error {
	client = GetClient()

	now := time.Now()
	updatedAtAttr, err := attributevalue.Marshal(now)
	if err != nil {
		return err
	}

	_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &tenantMembersTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: memberId},
		},
		UpdateExpression: aws.String("SET #S = :status, updatedAt = :updatedAt"),
		ExpressionAttributeNames: map[string]string{
			"#S": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":   &types.AttributeValueMemberS{Value: string(models.TenantMemberStatusRemoved)},
			":updatedAt": updatedAtAttr,
		},
	})
	return err
}

func GetTenantMemberByUserAndTenant(ctx context.Context, userId, tenantId string) (*models.TenantMember, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &tenantMembersTableName,
		IndexName:              aws.String("tenantUserIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid AND userId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: tenantId},
			":uid": &types.AttributeValueMemberS{Value: userId},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var member models.TenantMember
	err = attributevalue.UnmarshalMap(result.Items[0], &member)
	return &member, err
}

func IsTenantAdmin(ctx context.Context, userId, tenantId string) (bool, error) {
	member, err := GetTenantMemberByUserAndTenant(ctx, userId, tenantId)
	if err != nil {
		return false, err
	}
	return member != nil && member.Role == models.TenantMemberRoleAdmin && member.Status == models.TenantMemberStatusActive, nil
}

func UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	client = GetClient()
	tenant.UpdatedAt = time.Now()
	item, err := attributevalue.MarshalMap(tenant)
	if err != nil {
		return err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tenantsTableName,
		Item:      item,
	})
	return err
}

func GetTenantMemberById(ctx context.Context, memberId string) (*models.TenantMember, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tenantMembersTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: memberId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var member models.TenantMember
	err = attributevalue.UnmarshalMap(result.Item, &member)
	return &member, err
}
