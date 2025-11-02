package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TeamSettings struct {
	Id                          string    `dynamodbav:"id" json:"id"`
	TeamID                      string    `dynamodbav:"id" json:"teamId"`
	AllowFileUploads            bool      `dynamodbav:"allowFileUploads" json:"allowFileUploads"`
	AllowTeamGoalComments       bool      `dynamodbav:"allowTeamGoalComments" json:"allowTeamGoalComments"`
	AllowIndividualGoalComments bool      `dynamodbav:"allowIndividualGoalComments" json:"allowIndividualGoalComments"`
	CreatedAt                   time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt                   time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}

func (ts *TeamSettings) ToAttributeValues() map[string]types.AttributeValue {
	id, err := attributevalue.Marshal(ts.Id)
	if err != nil {
		return nil
	}
	teamID, err := attributevalue.Marshal(ts.TeamID)
	if err != nil {
		return nil
	}
	allowFileUploads, err := attributevalue.Marshal(ts.AllowFileUploads)
	if err != nil {
		return nil
	}
	allowTeamGoalComments, err := attributevalue.Marshal(ts.AllowTeamGoalComments)
	if err != nil {
		return nil
	}
	allowIndividualGoalComments, err := attributevalue.Marshal(ts.AllowIndividualGoalComments)
	if err != nil {
		return nil
	}
	createdAt, err := attributevalue.Marshal(ts.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return nil
	}
	updatedAt, err := attributevalue.Marshal(ts.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return nil
	}

	attributeValues := map[string]types.AttributeValue{
		"id":                          id,
		"teamId":                      teamID,
		"allowFileUploads":            allowFileUploads,
		"allowTeamGoalComments":       allowTeamGoalComments,
		"allowIndividualGoalComments": allowIndividualGoalComments,
		"createdAt":                   createdAt,
		"updatedAt":                   updatedAt,
	}
	return attributeValues
}
