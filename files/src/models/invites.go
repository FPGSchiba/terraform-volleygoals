package models

import "time"

type InviteStatus string

const (
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusRevoked  InviteStatus = "revoked"
	InviteStatusExpired  InviteStatus = "expired"
)

type Invite struct {
	Id         string         `dynamodbav:"id" json:"id"`
	TeamId     string         `dynamodbav:"teamId" json:"teamId"`
	Email      string         `dynamodbav:"email" json:"email"`
	Role       TeamMemberRole `dynamodbav:"role" json:"role"`
	Status     InviteStatus   `dynamodbav:"status" json:"status"`
	Token      string         `dynamodbav:"token" json:"token"`
	Message    *string        `dynamodbav:"message" json:"message"`
	InvitedBy  string         `dynamodbav:"invitedBy" json:"invitedBy"`
	AcceptedBy *string        `dynamodbav:"acceptedBy" json:"acceptedBy"`
	ExpiresAt  time.Time      `dynamodbav:"expiresAt" json:"expiresAt"`
	CreatedAt  time.Time      `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time      `dynamodbav:"updatedAt" json:"updatedAt"`
	AcceptedAt *time.Time     `dynamodbav:"acceptedAt" json:"acceptedAt"`
}
