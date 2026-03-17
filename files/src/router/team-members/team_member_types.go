package team_members

import (
	"time"

	"github.com/fpgschiba/volleygoals/models"
)

type TeamMemberListResult struct {
	Id                string                  `json:"id"`
	UserId            string                  `json:"userId"`
	Name              *string                 `json:"name"`
	Email             string                  `json:"email"`
	Picture           *string                 `json:"picture"`
	PreferredUsername *string                 `json:"preferredUsername"`
	Role              models.TeamMemberRole   `json:"role"`
	Status            models.TeamMemberStatus `json:"status"`
	UserStatus        models.UserStatus       `json:"userStatus"`
	Birthdate         *time.Time              `json:"birthdate"`
	JoinedAt          *time.Time              `json:"joinedAt,omitempty"`
}

type TeamMemberPublicResult struct {
	Id                string  `json:"id"`
	UserId            string  `json:"userId"`
	Name              *string `json:"name"`
	PreferredUsername *string `json:"preferredUsername"`
	Picture           *string `json:"picture"`
	Email             string  `json:"email"`
}

type AddTeamMemberRequest struct {
	UserId string                `json:"userId"`
	Role   models.TeamMemberRole `json:"role"`
}

type UpdateTeamMemberRequest struct {
	Role   *models.TeamMemberRole   `json:"role"`
	Status *models.TeamMemberStatus `json:"status"`
}
