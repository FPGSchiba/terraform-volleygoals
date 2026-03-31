package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type RoleDefinition struct {
	Id          string    `dynamodbav:"id" json:"id"`
	TenantId    string    `dynamodbav:"tenantId" json:"tenantId"` // "global" = applies to all tenants
	Name        string    `dynamodbav:"name" json:"name"`
	Permissions []string  `dynamodbav:"permissions" json:"permissions"`
	IsDefault   bool      `dynamodbav:"isDefault" json:"isDefault"`
	CreatedAt   time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}

func (r *RoleDefinition) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(r)
	if err != nil {
		return nil
	}
	return m
}
