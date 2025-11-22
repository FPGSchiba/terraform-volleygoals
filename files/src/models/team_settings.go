package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TeamSettings struct {
	Id                          string    `dynamodbav:"id" json:"id"`
	TeamID                      string    `dynamodbav:"teamId" json:"teamId"`
	AllowFileUploads            bool      `dynamodbav:"allowFileUploads" json:"allowFileUploads"`
	AllowTeamGoalComments       bool      `dynamodbav:"allowTeamGoalComments" json:"allowTeamGoalComments"`
	AllowIndividualGoalComments bool      `dynamodbav:"allowIndividualGoalComments" json:"allowIndividualGoalComments"`
	CreatedAt                   time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt                   time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}

func (t *TeamSettings) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(t)
	if err != nil {
		return nil
	}
	return m
}
