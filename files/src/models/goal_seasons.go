package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type GoalSeason struct {
	Id        string    `dynamodbav:"id" json:"id"`
	GoalId    string    `dynamodbav:"goalId" json:"goalId"`
	SeasonId  string    `dynamodbav:"seasonId" json:"seasonId"`
	CreatedAt time.Time `dynamodbav:"createdAt" json:"createdAt"`
}

func (gs *GoalSeason) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(gs)
	if err != nil {
		return nil
	}
	return m
}
