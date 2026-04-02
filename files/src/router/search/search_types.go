package search

// SearchGoalResult is a single goal hit in a search response.
type SearchGoalResult struct {
	Type   string `json:"type"`
	Id     string `json:"id"`
	Title  string `json:"title"`
	TeamId string `json:"teamId"`
	Status string `json:"status"`
}

// SearchReportResult is a single progress-report hit in a search response.
type SearchReportResult struct {
	Type      string `json:"type"`
	Id        string `json:"id"`
	Summary   string `json:"summary"`
	SeasonId  string `json:"seasonId"`
	CreatedAt string `json:"createdAt"`
}
