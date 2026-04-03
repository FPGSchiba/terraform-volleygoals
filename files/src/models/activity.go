package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ActivityVisibility string

const (
	ActivityVisibilityAll          ActivityVisibility = "all"
	ActivityVisibilityAdminTrainer ActivityVisibility = "admin_trainer"
)

type Activity struct {
	Id           string             `dynamodbav:"id" json:"id"`
	TeamId       string             `dynamodbav:"teamId" json:"teamId"`
	ActorId      string             `dynamodbav:"actorId" json:"actorId"`
	ActorName    string             `dynamodbav:"actorName" json:"actorName"`
	ActorPicture string             `dynamodbav:"actorPicture" json:"actorPicture,omitempty"`
	Action       string             `dynamodbav:"action" json:"action"`
	Description  string             `dynamodbav:"description" json:"description"`
	TargetType   string             `dynamodbav:"targetType" json:"targetType,omitempty"`
	TargetId      string             `dynamodbav:"targetId" json:"targetId,omitempty"`
	TargetOwnerId string             `dynamodbav:"targetOwnerId" json:"targetOwnerId,omitempty"`
	Visibility    ActivityVisibility `dynamodbav:"visibility" json:"visibility"`
	Timestamp    time.Time          `dynamodbav:"timestamp" json:"timestamp"`
}

func (a *Activity) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(a)
	if err != nil {
		return nil
	}
	return m
}
