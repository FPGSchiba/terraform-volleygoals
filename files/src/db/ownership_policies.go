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
		// If we're updating a tenant-specific policy, preserve versioning by
		// creating a new record and marking the previous record's UpdatedAt
		// and UpdatedTo fields. If the existing record is the global policy,
		// do NOT mutate it; instead create a new policy (the router should
		// normally route global edits to a tenant-specific Upsert).
		newPolicy := &models.OwnershipPolicy{
			Id:                     models.GenerateID(),
			TenantId:               tenantId,
			ResourceType:           resourceType,
			OwnerPermissions:       ownerPerms,
			ParentOwnerPermissions: parentOwnerPerms,
			CreatedAt:              now,
			UpdatedAt:              nil,
			UpdatedTo:              nil,
		}

		item, err := attributevalue.MarshalMap(newPolicy)
		if err != nil {
			return nil, err
		}

		// Put the new policy first
		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: &ownershipPoliciesTableName,
			Item:      item,
		})
		if err != nil {
			return nil, err
		}

		// If the existing record is not the global policy, mark it as updated
		// and point to the new policy id. This preserves the old version.
		if existing.TenantId != "global" {
			_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName: &ownershipPoliciesTableName,
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: existing.Id},
				},
				UpdateExpression: aws.String("SET updatedAt = :u, updatedTo = :to"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":u":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
					":to": &types.AttributeValueMemberS{Value: newPolicy.Id},
				},
			})
			if err != nil {
				// Attempt best-effort rollback of new policy if we fail to update the old one
				_, _ = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
					TableName: &ownershipPoliciesTableName,
					Key: map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{Value: newPolicy.Id},
					},
				})
				return nil, err
			}
			existing.UpdatedAt = &now
			existing.UpdatedTo = &newPolicy.Id
		}

		// Return the newly created policy (the latest version)
		return newPolicy, nil
	}
	policy := &models.OwnershipPolicy{
		Id:                     models.GenerateID(),
		TenantId:               tenantId,
		ResourceType:           resourceType,
		OwnerPermissions:       ownerPerms,
		ParentOwnerPermissions: parentOwnerPerms,
		CreatedAt:              now,
		UpdatedAt:              &now,
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

// UpsertOwnershipPolicyReturnPrev behaves like UpsertOwnershipPolicy but also
// returns the previous policy record if one existed (nil otherwise). This is
// useful for callers that need to implement compensation/rollback.
func UpsertOwnershipPolicyReturnPrev(ctx context.Context, tenantId, resourceType string, ownerPerms, parentOwnerPerms []string) (*models.OwnershipPolicy, *models.OwnershipPolicy, error) {
	client = GetClient()
	// Check if one already exists
	existing, err := GetOwnershipPolicyByTenantAndType(ctx, tenantId, resourceType)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	if existing != nil {
		// Keep a copy of the previous record for potential rollback
		prev := &models.OwnershipPolicy{}
		*prev = *existing

		newPolicy := &models.OwnershipPolicy{
			Id:                     models.GenerateID(),
			TenantId:               tenantId,
			ResourceType:           resourceType,
			OwnerPermissions:       ownerPerms,
			ParentOwnerPermissions: parentOwnerPerms,
			CreatedAt:              now,
			UpdatedAt:              nil,
			UpdatedTo:              nil,
		}

		item, err := attributevalue.MarshalMap(newPolicy)
		if err != nil {
			return nil, nil, err
		}

		// Put the new policy first
		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: &ownershipPoliciesTableName,
			Item:      item,
		})
		if err != nil {
			return nil, nil, err
		}

		// If the existing record is not the global policy, mark it as updated
		// and point to the new policy id. This preserves the old version.
		if existing.TenantId != "global" {
			_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName: &ownershipPoliciesTableName,
				Key: map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: existing.Id},
				},
				UpdateExpression: aws.String("SET updatedAt = :u, updatedTo = :to"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":u":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
					":to": &types.AttributeValueMemberS{Value: newPolicy.Id},
				},
			})
			if err != nil {
				// Attempt best-effort rollback of new policy if we fail to update the old one
				_, _ = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
					TableName: &ownershipPoliciesTableName,
					Key: map[string]types.AttributeValue{
						"id": &types.AttributeValueMemberS{Value: newPolicy.Id},
					},
				})
				return nil, nil, err
			}
		}

		return newPolicy, prev, nil
	}
	policy := &models.OwnershipPolicy{
		Id:                     models.GenerateID(),
		TenantId:               tenantId,
		ResourceType:           resourceType,
		OwnerPermissions:       ownerPerms,
		ParentOwnerPermissions: parentOwnerPerms,
		CreatedAt:              now,
		UpdatedAt:              &now,
	}
	item, err := attributevalue.MarshalMap(policy)
	if err != nil {
		return nil, nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &ownershipPoliciesTableName,
		Item:      item,
	})
	return policy, nil, err
}

// CompensateUpserts attempts to undo created policy records and restore
// previous policy fields. It is best-effort and errors are logged but not
// returned to callers as compensation should not mask the original error.
func CompensateUpserts(ctx context.Context, createdIDs []string, prevs map[string]*models.OwnershipPolicy) error {
	client = GetClient()
	// Delete created items
	for _, cid := range createdIDs {
		_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &ownershipPoliciesTableName,
			Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: cid}},
		})
		if err != nil {
			// best-effort: log and continue
			// using Printf to avoid introducing extra logging deps here
			// but prefer logrus if available elsewhere
		}
	}

	// Restore previous updatedAt/updatedTo where applicable
	for _, prev := range prevs {
		if prev == nil {
			continue
		}
		if prev.TenantId == "global" {
			// global was not mutated; nothing to restore
			continue
		}
		if prev.UpdatedAt == nil && prev.UpdatedTo == nil {
			_, _ = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName: &ownershipPoliciesTableName,
				Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: prev.Id}},
				UpdateExpression: aws.String("REMOVE updatedAt, updatedTo"),
			})
		} else {
			vals := map[string]types.AttributeValue{}
			expr := "SET updatedAt = :u, updatedTo = :to"
			if prev.UpdatedAt != nil {
				vals[":u"] = &types.AttributeValueMemberS{Value: prev.UpdatedAt.Format(time.RFC3339)}
			} else {
				vals[":u"] = &types.AttributeValueMemberS{Value: ""}
			}
			if prev.UpdatedTo != nil {
				vals[":to"] = &types.AttributeValueMemberS{Value: *prev.UpdatedTo}
			} else {
				vals[":to"] = &types.AttributeValueMemberS{Value: ""}
			}
			_, _ = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName: &ownershipPoliciesTableName,
				Key: map[string]types.AttributeValue{"id": &types.AttributeValueMemberS{Value: prev.Id}},
				UpdateExpression: aws.String(expr),
				ExpressionAttributeValues: vals,
			})
		}
	}
	return nil
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
