package progress_reports

import "github.com/fpgschiba/volleygoals/models"

type ProgressReportWithProgress struct {
	*models.ProgressReport
	Progress []*models.Progress `json:"progress"`
}

type ProgressEntry struct {
	GoalId  string `json:"goalId"`
	Rating  int8   `json:"rating"`
	Details string `json:"details,omitempty"`
}

type CreateProgressReportRequest struct {
	Summary        string          `json:"summary"`
	Details        string          `json:"details"`
	OverallDetails string          `json:"overallDetails"`
	Progress       []ProgressEntry `json:"progress,omitempty"`
}

type UpdateProgressReportRequest struct {
	Summary        *string         `json:"summary,omitempty"`
	Details        *string         `json:"details,omitempty"`
	OverallDetails *string         `json:"overallDetails,omitempty"`
	Progress       []ProgressEntry `json:"progress,omitempty"`
}
