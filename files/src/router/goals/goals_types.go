package goals

import "github.com/fpgschiba/volleygoals/models"

type CreateGoalRequest struct {
	Type        models.GoalType `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	OwnerId     *string         `json:"ownerId,omitempty"`
}

type UpdateGoalRequest struct {
	OwnerId     *string            `json:"ownerId,omitempty"`
	Title       *string            `json:"title,omitempty"`
	Description *string            `json:"description,omitempty"`
	Status      *models.GoalStatus `json:"status,omitempty"`
}

type GoalOwner struct {
	Id                string  `json:"id"`
	Name              *string `json:"name"`
	PreferredUsername *string `json:"preferredUsername"`
	Picture           *string `json:"picture"`
}

type GoalWithOwner struct {
	*models.Goal
	Owner                *GoalOwner `json:"owner,omitempty"`
	CompletionPercentage int        `json:"completionPercentage"`
}
