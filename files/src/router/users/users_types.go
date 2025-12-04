package users

import "github.com/fpgschiba/volleygoals/models"

type UpdateUserRequest struct {
	UserType *models.UserType `json:"userType"`
	Enabled  *bool            `json:"enabled"`
}
