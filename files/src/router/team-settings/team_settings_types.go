package team_settings

type UpdateTeamSettingsRequest struct {
	AllowFileUploads            *bool `json:"allowFileUploads"`
	AllowTeamGoalComments       *bool `json:"allowTeamGoalComments"`
	AllowIndividualGoalComments *bool `json:"allowIndividualGoalComments"`
}
