package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TeamStatus string

const (
	TeamStatusActive   TeamStatus = "active"
	TeamStatusInactive TeamStatus = "inactive"
)

type Team struct {
	Id        string     `dynamodbav:"id" json:"id"`
	Name      string     `dynamodbav:"teamName" json:"name"`
	Status    TeamStatus `dynamodbav:"status" json:"status"`
	Picture   string     `dynamodbav:"picture" json:"picture"`
	CreatedAt time.Time  `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time  `dynamodbav:"updatedAt" json:"updatedAt"`
	DeletedAt *time.Time `dynamodbav:"deletedAt" json:"deletedAt"`
}

func (t *Team) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(t)
	if err != nil {
		return nil
	}
	return m
}
