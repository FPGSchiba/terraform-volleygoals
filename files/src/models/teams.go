package models

import "time"

type TeamStatus string

const (
	TeamStatusActive   TeamStatus = "active"
	TeamStatusInactive TeamStatus = "inactive"
)

type Team struct {
	Id        string     `dynamodbav:"id" json:"id"`
	Name      string     `dynamodbav:"name" json:"name"`
	Status    TeamStatus `dynamodbav:"status" json:"status"`
	CreatedAt time.Time  `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time  `dynamodbav:"updatedAt" json:"updatedAt"`
	DeletedAt *time.Time `dynamodbav:"deletedAt" json:"deletedAt"`
}
