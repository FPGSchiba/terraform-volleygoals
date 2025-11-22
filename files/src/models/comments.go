package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type CommentType string

const (
	CommentTypeProgressReport CommentType = "ProgressReport"
	CommentTypeGoal           CommentType = "Goal"
)

type Comment struct {
	Id        string      `dynamodbav:"id" json:"id"`
	AuthorId  string      `dynamodbav:"authorId" json:"authorId"`
	Type      CommentType `dynamodbav:"type" json:"type"`
	TargetId  string      `dynamodbav:"targetId" json:"targetId"`
	Content   string      `dynamodbav:"content" json:"content"`
	CreatedAt time.Time   `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time   `dynamodbav:"updatedAt" json:"updatedAt"`
}

func (c *Comment) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(c)
	if err != nil {
		return nil
	}
	return m
}
