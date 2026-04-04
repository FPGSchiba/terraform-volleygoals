package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ProgressReport struct {
	Id             string    `dynamodbav:"id" json:"id"`
	SeasonId       string    `dynamodbav:"seasonId" json:"seasonId"`
	AuthorId       string    `dynamodbav:"authorId" json:"authorId"`
	Summary        string    `dynamodbav:"summary" json:"summary"`
	Details        string    `dynamodbav:"details" json:"details"`
	OverallDetails string    `dynamodbav:"overallDetails" json:"overallDetails"`
	AuthorName     *string   `dynamodbav:"authorName,omitempty" json:"authorName,omitempty"`
	AuthorPicture  *string   `dynamodbav:"authorPicture,omitempty" json:"authorPicture,omitempty"`
	ReportDate     time.Time `dynamodbav:"reportDate" json:"reportDate"` // TODO: Implement this so the user can set the date when they did this (so one can create reports after the fact)
	CreatedAt      time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}

func (p *ProgressReport) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(p)
	if err != nil {
		return nil
	}
	return m
}
