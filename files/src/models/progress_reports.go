package models

import "time"

type ProgressReport struct {
	Id        string    `dynamodbav:"id" json:"id"`
	SeasonId  string    `dynamodbav:"seasonId" json:"seasonId"`
	AuthorId  string    `dynamodbav:"authorId" json:"authorId"`
	Summary   string    `dynamodbav:"summary" json:"summary"`
	Details   string    `dynamodbav:"details" json:"details"`
	CreatedAt time.Time `dynamodbav:"createdAt" json:"createdAt"`
}
