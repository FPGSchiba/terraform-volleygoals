package invites

import "github.com/fpgschiba/volleygoals/models"

type CreateInviteRequest struct {
	Email     string                `json:"email"`
	TeamId    string                `json:"teamId"`
	Role      models.TeamMemberRole `json:"role"`
	SendEmail bool                  `json:"sendEmail"`
	Message   *string               `json:"message"` // TODO: Actually use this in the mail and creation of the invite
}

type CompleteInviteRequest struct {
	Token    string `json:"token"`
	Email    string `json:"email"`
	Accepted bool   `json:"accepted"`
}
