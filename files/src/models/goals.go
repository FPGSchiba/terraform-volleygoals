package models

import "time"

type GoalType string

type GoalStatus string

const (
	GoalTypeIndividual GoalType = "individual"
	GoalTypeTeam       GoalType = "team"

	GoalStatusOpen       GoalStatus = "open"
	GoalStatusInProgress GoalStatus = "in_progress"
	GoalStatusCompleted  GoalStatus = "completed"
	GoalStatusArchived   GoalStatus = "archived"
)

type Goal struct {
	Id          string     `dynamodbav:"id" json:"id"`
	SeasonId    string     `dynamodbav:"seasonId" json:"seasonId"`
	OwnerId     string     `dynamodbav:"ownerId" json:"ownerId"`
	Type        GoalType   `dynamodbav:"type" json:"type"`
	Title       string     `dynamodbav:"title" json:"title"`
	Description string     `dynamodbav:"description" json:"description"`
	Status      GoalStatus `dynamodbav:"status" json:"status"`
	CreatedBy   string     `dynamodbav:"createdBy" json:"createdBy"`
	CreatedAt   time.Time  `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time  `dynamodbav:"updatedAt" json:"updatedAt"`
}
