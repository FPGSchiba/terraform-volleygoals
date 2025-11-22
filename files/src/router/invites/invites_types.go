package invites

import "github.com/fpgschiba/volleygoals/models"

type CreateInviteRequest struct {
	Email  string                `json:"email"`
	TeamId string                `json:"teamId"`
	Role   models.TeamMemberRole `json:"role"`
}
