package progress_reports

type ProgressEntry struct {
	GoalId string `json:"goalId"`
	Rating int8   `json:"rating"`
}

type CreateProgressReportRequest struct {
	Summary  string          `json:"summary"`
	Details  string          `json:"details"`
	Progress []ProgressEntry `json:"progress,omitempty"`
}

type UpdateProgressReportRequest struct {
	Summary  *string         `json:"summary,omitempty"`
	Details  *string         `json:"details,omitempty"`
	Progress []ProgressEntry `json:"progress,omitempty"`
}
