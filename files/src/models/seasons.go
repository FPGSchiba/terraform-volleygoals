package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type SeasonStatus string

const (
	SeasonStatusPlanned   SeasonStatus = "planned"
	SeasonStatusActive    SeasonStatus = "active"
	SeasonStatusCompleted SeasonStatus = "completed"
	SeasonStatusArchived  SeasonStatus = "archived"
)

type Season struct {
	Id        string       `dynamodbav:"id" json:"id"`
	TeamId    string       `dynamodbav:"teamId" json:"teamId"`
	Name      string       `dynamodbav:"name" json:"name"`
	StartDate time.Time    `dynamodbav:"startDate" json:"startDate"`
	EndDate   time.Time    `dynamodbav:"endDate" json:"endDate"`
	Status    SeasonStatus `dynamodbav:"status" json:"status"`
	CreatedAt time.Time    `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time    `dynamodbav:"updatedAt" json:"updatedAt"`
}

func (s *Season) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(s)
	if err != nil {
		return nil
	}
	return m
}
