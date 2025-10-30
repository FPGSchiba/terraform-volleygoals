package models

import "time"

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
	TeamId     string           `dynamodbav:"teamId" json:"teamId"`
	CognitoSub string           `dynamodbav:"cognitoSub" json:"cognitoSub"`
	Role       TeamMemberRole   `dynamodbav:"role" json:"role"`
	Status     TeamMemberStatus `dynamodbav:"status" json:"status"`
	CreatedAt  time.Time        `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time        `dynamodbav:"updatedAt" json:"updatedAt"`
	JoinedAt   *time.Time       `dynamodbav:"joinedAt" json:"joinedAt"`
	LeftAt     *time.Time       `dynamodbav:"leftAt" json:"leftAt"`
}
