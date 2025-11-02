package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TeamStatus string

const (
	TeamStatusActive   TeamStatus = "active"
	TeamStatusInactive TeamStatus = "inactive"
)

type Team struct {
	Id        string     `dynamodbav:"id" json:"id"`
	Name      string     `dynamodbav:"teamName" json:"name"`
	Status    TeamStatus `dynamodbav:"status" json:"status"`
	CreatedAt time.Time  `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time  `dynamodbav:"updatedAt" json:"updatedAt"`
	DeletedAt *time.Time `dynamodbav:"deletedAt" json:"deletedAt"`
}

func (t *Team) ToAttributeValues() map[string]types.AttributeValue {
	id, err := attributevalue.Marshal(t.Id)
	if err != nil {
		return nil
	}
	name, err := attributevalue.Marshal(t.Name)
	if err != nil {
		return nil
	}
	status, err := attributevalue.Marshal(t.Status)
	if err != nil {
		return nil
	}
	createdAt, err := attributevalue.Marshal(t.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return nil
	}
	updatedAt, err := attributevalue.Marshal(t.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return nil
	}

	attributeValues := map[string]types.AttributeValue{
		"id":        id,
		"teamName":  name,
		"status":    status,
		"createdAt": createdAt,
		"updatedAt": updatedAt,
	}

	if t.DeletedAt != nil {
		deletedAt, err := attributevalue.Marshal(t.DeletedAt.Format(time.RFC3339))
		if err != nil {
			return nil
		}
		attributeValues["deletedAt"] = deletedAt
	}

	return attributeValues
}
