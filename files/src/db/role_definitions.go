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

// GetRoleDefinitionByTenantAndName returns the role definition for the given
// tenantId and role name. If no tenant-specific definition exists it falls back
// to the global default (tenantId = "global"). Returns nil if not found.
func GetRoleDefinitionByTenantAndName(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
	client = GetClient()
	candidates := []string{tenantId, "global"}
	for _, tid := range candidates {
		if tid == "" {
			continue
		}
		result, err := client.Query(ctx, &dynamodb.QueryInput{
			TableName:              &roleDefinitionsTableName,
			IndexName:              aws.String("tenantNameIndex"),
			KeyConditionExpression: aws.String("tenantId = :tid AND #name = :name"),
			ExpressionAttributeNames: map[string]string{
				"#name": "name",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":tid":  &types.AttributeValueMemberS{Value: tid},
				":name": &types.AttributeValueMemberS{Value: roleName},
			},
			Limit: aws.Int32(1),
		})
		if err != nil {
			return nil, err
		}
		if len(result.Items) > 0 {
			var def models.RoleDefinition
			if err := attributevalue.UnmarshalMap(result.Items[0], &def); err != nil {
				return nil, err
			}
			return &def, nil
		}
	}
	return nil, nil
}

// GetRoleDefinitionByTenantExact returns the role definition for the given
// tenantId and role name with NO global fallback. Returns nil if not found.
func GetRoleDefinitionByTenantExact(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
	client = GetClient()
	if tenantId == "" {
		return nil, nil
	}
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &roleDefinitionsTableName,
		IndexName:              aws.String("tenantNameIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid AND #name = :name"),
		ExpressionAttributeNames: map[string]string{
			"#name": "name",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid":  &types.AttributeValueMemberS{Value: tenantId},
			":name": &types.AttributeValueMemberS{Value: roleName},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("GetRoleDefinitionByTenantExact: query: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var def models.RoleDefinition
	if err := attributevalue.UnmarshalMap(result.Items[0], &def); err != nil {
		return nil, fmt.Errorf("GetRoleDefinitionByTenantExact: unmarshal: %w", err)
	}
	return &def, nil
}

func ListRoleDefinitionsByTenant(ctx context.Context, tenantId string) ([]*models.RoleDefinition, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &roleDefinitionsTableName,
		IndexName:              aws.String("tenantIdIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: tenantId},
		},
	})
	if err != nil {
		return nil, err
	}
	var defs []*models.RoleDefinition
	err = attributevalue.UnmarshalListOfMaps(result.Items, &defs)
	return defs, err
}

func CreateRoleDefinition(ctx context.Context, tenantId, name string, permissions []string, isDefault bool) (*models.RoleDefinition, error) {
	client = GetClient()
	now := time.Now()
	def := &models.RoleDefinition{
		Id:          models.GenerateID(),
		TenantId:    tenantId,
		Name:        name,
		Permissions: permissions,
		IsDefault:   isDefault,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	item, err := attributevalue.MarshalMap(def)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &roleDefinitionsTableName,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	return def, nil
}

func UpdateRoleDefinitionPermissions(ctx context.Context, roleId string, permissions []string) (*models.RoleDefinition, error) {
	client = GetClient()
	permList, err := attributevalue.MarshalList(permissions)
	if err != nil {
		return nil, err
	}
	result, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &roleDefinitionsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: roleId},
		},
		UpdateExpression: aws.String("SET permissions = :p, updatedAt = :u"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberL{Value: permList},
			":u": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		return nil, err
	}
	var def models.RoleDefinition
	err = attributevalue.UnmarshalMap(result.Attributes, &def)
	return &def, err
}

func DeleteRoleDefinition(ctx context.Context, roleId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &roleDefinitionsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: roleId},
		},
	})
	return err
}

func GetRoleDefinitionById(ctx context.Context, roleId string) (*models.RoleDefinition, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &roleDefinitionsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: roleId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var def models.RoleDefinition
	err = attributevalue.UnmarshalMap(result.Item, &def)
	return &def, err
}
