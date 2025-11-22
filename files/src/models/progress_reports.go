package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ProgressReport struct {
	Id        string    `dynamodbav:"id" json:"id"`
	SeasonId  string    `dynamodbav:"seasonId" json:"seasonId"`
	AuthorId  string    `dynamodbav:"authorId" json:"authorId"`
	Summary   string    `dynamodbav:"summary" json:"summary"`
	Details   string    `dynamodbav:"details" json:"details"`
	CreatedAt time.Time `dynamodbav:"createdAt" json:"createdAt"`
}

func (p *ProgressReport) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(p)
	if err != nil {
		return nil
	}
	return m
}
