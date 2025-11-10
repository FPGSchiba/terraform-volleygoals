package teams

import "github.com/fpgschiba/volleygoals/models"

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type UpdateTeamRequest struct {
	Name   string            `json:"name"`
	Status models.TeamStatus `json:"status"`
}
