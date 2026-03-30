package models

import "time"

type OwnershipPolicy struct {
	Id                     string    `dynamodbav:"id" json:"id"`
	TenantId               string    `dynamodbav:"tenantId" json:"tenantId"` // "global" = applies to all tenants
	ResourceType           string    `dynamodbav:"resourceType" json:"resourceType"`
	OwnerPermissions       []string  `dynamodbav:"ownerPermissions" json:"ownerPermissions"`
	ParentOwnerPermissions []string  `dynamodbav:"parentOwnerPermissions" json:"parentOwnerPermissions"`
	CreatedAt              time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt              time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}
