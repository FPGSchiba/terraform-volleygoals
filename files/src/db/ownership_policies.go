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

// GetOwnershipPolicy returns the ownership policy for the given tenantId and
// resourceType. Falls back to the global default (tenantId = "global").
func GetOwnershipPolicy(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error) {
	client = GetClient()
	candidates := []string{tenantId, "global"}
	for _, tid := range candidates {
		if tid == "" {
			continue
		}
		result, err := client.Query(ctx, &dynamodb.QueryInput{
			TableName:              &ownershipPoliciesTableName,
			IndexName:              aws.String("tenantResourceTypeIndex"),
			KeyConditionExpression: aws.String("tenantId = :tid AND resourceType = :rt"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":tid": &types.AttributeValueMemberS{Value: tid},
				":rt":  &types.AttributeValueMemberS{Value: resourceType},
			},
			Limit: aws.Int32(1),
		})
		if err != nil {
			return nil, err
		}
		if len(result.Items) > 0 {
			var policy models.OwnershipPolicy
			if err := attributevalue.UnmarshalMap(result.Items[0], &policy); err != nil {
				return nil, err
			}
			return &policy, nil
		}
	}
	return nil, nil
}

func ListOwnershipPoliciesByTenant(ctx context.Context, tenantId string) ([]*models.OwnershipPolicy, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &ownershipPoliciesTableName,
		IndexName:              aws.String("tenantIdIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: tenantId},
		},
	})
	if err != nil {
		return nil, err
	}
	var policies []*models.OwnershipPolicy
	err = attributevalue.UnmarshalListOfMaps(result.Items, &policies)
	return policies, err
}

func UpsertOwnershipPolicy(ctx context.Context, tenantId, resourceType string, ownerPerms, parentOwnerPerms []string) (*models.OwnershipPolicy, error) {
	client = GetClient()
	// Check if one already exists
	existing, err := GetOwnershipPolicyByTenantAndType(ctx, tenantId, resourceType)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if existing != nil {
		ownerList, err := attributevalue.MarshalList(ownerPerms)
		if err != nil {
			return nil, err
		}
		parentList, err := attributevalue.MarshalList(parentOwnerPerms)
		if err != nil {
			return nil, err
		}
		_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: &ownershipPoliciesTableName,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: existing.Id},
			},
			UpdateExpression: aws.String("SET ownerPermissions = :op, parentOwnerPermissions = :pp, updatedAt = :u"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":op": &types.AttributeValueMemberL{Value: ownerList},
				":pp": &types.AttributeValueMemberL{Value: parentList},
				":u":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			},
		})
		if err != nil {
			return nil, err
		}
		existing.OwnerPermissions = ownerPerms
		existing.ParentOwnerPermissions = parentOwnerPerms
		existing.UpdatedAt = now
		return existing, nil
	}
	policy := &models.OwnershipPolicy{
		Id:                     models.GenerateID(),
		TenantId:               tenantId,
		ResourceType:           resourceType,
		OwnerPermissions:       ownerPerms,
		ParentOwnerPermissions: parentOwnerPerms,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	item, err := attributevalue.MarshalMap(policy)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &ownershipPoliciesTableName,
		Item:      item,
	})
	return policy, err
}

// GetOwnershipPolicyByTenantAndType fetches the exact record for a given
// tenantId (no fallback). Used internally by UpsertOwnershipPolicy.
func GetOwnershipPolicyByTenantAndType(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &ownershipPoliciesTableName,
		IndexName:              aws.String("tenantResourceTypeIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid AND resourceType = :rt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: tenantId},
			":rt":  &types.AttributeValueMemberS{Value: resourceType},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var policy models.OwnershipPolicy
	err = attributevalue.UnmarshalMap(result.Items[0], &policy)
	return &policy, err
}
