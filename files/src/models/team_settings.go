package models

import "time"

type TeamSettings struct {
	Id                          string    `dynamodbav:"id" json:"id"`
	TeamID                      string    `dynamodbav:"id" json:"teamId"`
	AllowFileUploads            bool      `dynamodbav:"allowFileUploads" json:"allowFileUploads"`
	AllowTeamGoalComments       bool      `dynamodbav:"allowTeamGoalComments" json:"allowTeamGoalComments"`
	AllowIndividualGoalComments bool      `dynamodbav:"allowIndividualGoalComments" json:"allowIndividualGoalComments"`
	CreatedAt                   time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt                   time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}
