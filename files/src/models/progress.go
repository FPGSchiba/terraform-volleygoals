package models

type Progress struct {
	Id               string `dynamodbav:"id" json:"id"`
	ProgressReportId string `dynamodbav:"progressReportId" json:"progressReportId"`
	GoalId           string `dynamodbav:"goalId" json:"goalId"`
	Rating           int8   `dynamodbav:"rating" json:"rating"`
}
