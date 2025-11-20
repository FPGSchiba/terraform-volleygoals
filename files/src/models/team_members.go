package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TeamMemberRole string
type TeamMemberStatus string

const (
	TeamMemberRoleAdmin   TeamMemberRole = "admin"
	TeamMemberRoleTrainer TeamMemberRole = "trainer"
	TeamMemberRoleMember  TeamMemberRole = "member"

	TeamMemberStatusActive  TeamMemberStatus = "active"
	TeamMemberStatusInvited TeamMemberStatus = "invited"
	TeamMemberStatusRemoved TeamMemberStatus = "removed"
	TeamMemberStatusLeft    TeamMemberStatus = "left"
)

type TeamMember struct {
	Id         string           `dynamodbav:"id" json:"id"`
	CognitoSub string           `dynamodbav:"cognitoSub" json:"cognitoSub"`
	TeamId     string           `dynamodbav:"teamId" json:"teamId"`
	Role       TeamMemberRole   `dynamodbav:"role" json:"role"`
	Status     TeamMemberStatus `dynamodbav:"status" json:"status"`
	CreatedAt  time.Time        `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time        `dynamodbav:"updatedAt" json:"updatedAt"`
	JoinedAt   *time.Time       `dynamodbav:"joinedAt" json:"joinedAt"`
	LeftAt     *time.Time       `dynamodbav:"leftAt" json:"leftAt"`
}

func (tm *TeamMember) ToAttributeValues() map[string]types.AttributeValue {
	id, err := attributevalue.Marshal(tm.Id)
	if err != nil {
		return nil
	}
	cognitoSub, err := attributevalue.Marshal(tm.CognitoSub)
	if err != nil {
		return nil
	}
	teamId, err := attributevalue.Marshal(tm.TeamId)
	if err != nil {
		return nil
	}
	role, err := attributevalue.Marshal(tm.Role)
	if err != nil {
		return nil
	}
	status, err := attributevalue.Marshal(tm.Status)
	if err != nil {
		return nil
	}
	createdAt, err := attributevalue.Marshal(tm.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return nil
	}
	updatedAt, err := attributevalue.Marshal(tm.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return nil
	}

	attributeValues := map[string]types.AttributeValue{
		"id":         id,
		"cognitoSub": cognitoSub,
		"teamId":     teamId,
		"role":       role,
		"status":     status,
		"createdAt":  createdAt,
		"updatedAt":  updatedAt,
	}
	if tm.JoinedAt != nil {
		joinedAt, err := attributevalue.Marshal(tm.JoinedAt.Format(time.RFC3339))
		if err == nil {
			attributeValues["joinedAt"] = joinedAt
		}

	}
	if tm.LeftAt != nil {
		leftAt, err := attributevalue.Marshal(tm.LeftAt.Format(time.RFC3339))
		if err == nil {
			attributeValues["leftAt"] = leftAt
		}
	}
	return attributeValues
}
