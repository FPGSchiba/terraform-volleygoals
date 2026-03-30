package models

import "time"

type RoleDefinition struct {
	Id          string    `dynamodbav:"id" json:"id"`
	TenantId    string    `dynamodbav:"tenantId" json:"tenantId"` // "global" = applies to all tenants
	Name        string    `dynamodbav:"name" json:"name"`
	Permissions []string  `dynamodbav:"permissions" json:"permissions"`
	IsDefault   bool      `dynamodbav:"isDefault" json:"isDefault"`
	CreatedAt   time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}
