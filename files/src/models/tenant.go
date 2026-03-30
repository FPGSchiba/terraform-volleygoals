package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TenantMemberRole   string
type TenantMemberStatus string

const (
	TenantMemberRoleAdmin  TenantMemberRole = "tenant_admin"
	TenantMemberRoleMember TenantMemberRole = "tenant_member"

	TenantMemberStatusActive  TenantMemberStatus = "active"
	TenantMemberStatusRemoved TenantMemberStatus = "removed"
)

type Tenant struct {
	Id        string    `dynamodbav:"id" json:"id"`
	Name      string    `dynamodbav:"name" json:"name"`
	OwnerId   string    `dynamodbav:"ownerId" json:"ownerId"`
	CreatedAt time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}

type TenantMember struct {
	Id        string             `dynamodbav:"id" json:"id"`
	TenantId  string             `dynamodbav:"tenantId" json:"tenantId"`
	UserId    string             `dynamodbav:"userId" json:"userId"`
	Role      TenantMemberRole   `dynamodbav:"role" json:"role"`
	Status    TenantMemberStatus `dynamodbav:"status" json:"status"`
	CreatedAt time.Time          `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `dynamodbav:"updatedAt" json:"updatedAt"`
}

func (t *Tenant) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(t)
	if err != nil {
		return nil
	}
	return m
}

func (tm *TenantMember) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(tm)
	if err != nil {
		return nil
	}
	return m
}
