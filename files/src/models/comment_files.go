package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type CommentFile struct {
	Id         string    `dynamodbav:"id" json:"id"`
	CommentId  string    `dynamodbav:"commentId" json:"commentId"`
	StorageKey string    `dynamodbav:"storageKey" json:"storageKey"`
	Filename   string    `dynamodbav:"filename" json:"filename"`
	CreatedAt  time.Time `dynamodbav:"createdAt" json:"createdAt"`
}

func (c *CommentFile) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(c)
	if err != nil {
		return nil
	}
	return m
}
