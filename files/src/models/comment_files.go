package models

import "time"

type CommentFile struct {
	Id         string    `dynamodbav:"id" json:"id"`
	CommentId  string    `dynamodbav:"commentId" json:"commentId"`
	StorageKey string    `dynamodbav:"storageKey" json:"storageKey"`
	Filename   string    `dynamodbav:"filename" json:"filename"`
	CreatedAt  time.Time `dynamodbav:"createdAt" json:"createdAt"`
}
