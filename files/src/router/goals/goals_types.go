package goals

import "github.com/fpgschiba/volleygoals/models"

type CreateGoalRequest struct {
	Type        models.GoalType `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
}

type UpdateGoalRequest struct {
	OwnerId     *string            `json:"ownerId,omitempty"`
	Title       *string            `json:"title,omitempty"`
	Description *string            `json:"description,omitempty"`
	Status      *models.GoalStatus `json:"status,omitempty"`
}
