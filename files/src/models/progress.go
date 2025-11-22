package models

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type Progress struct {
	Id               string `dynamodbav:"id" json:"id"`
	ProgressReportId string `dynamodbav:"progressReportId" json:"progressReportId"`
	GoalId           string `dynamodbav:"goalId" json:"goalId"`
	Rating           int8   `dynamodbav:"rating" json:"rating"`
}

func (p *Progress) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(p)
	if err != nil {
		return nil
	}
	return m
}
