package seasons

import (
	"time"

	"github.com/fpgschiba/volleygoals/models"
)

type CreateSeasonRequest struct {
	TeamId    string    `json:"teamId"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type UpdateSeasonRequest struct {
	Name      *string              `json:"name,omitempty"`
	StartDate *time.Time           `json:"startDate,omitempty"`
	EndDate   *time.Time           `json:"endDate,omitempty"`
	Status    *models.SeasonStatus `json:"status,omitempty"`
}
