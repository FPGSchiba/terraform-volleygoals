//go:build !local

package db

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Table Names
var (
	teamsTableName           = os.Getenv("TEAMS_TABLE_NAME")
	invitesTableName         = os.Getenv("TEAM_MEMBERS_TABLE_NAME")
	teamMembersTableName     = os.Getenv("INVITE_TABLE_NAME")
	teamSettingsTableName    = os.Getenv("TEAM_SETTINGS_TABLE_NAME")
	seasonsTableName         = os.Getenv("SEASONS_TABLE_NAME")
	goalsTableName           = os.Getenv("GOALS_TABLE_NAME")
	progressReportsTableName = os.Getenv("PROGRESS_REPORTS_TABLE_NAME")
	progressTableName        = os.Getenv("PROGRESS_TABLE_NAME")
	commentsTableName        = os.Getenv("COMMENTS_TABLE_NAME")
	commentFilesTableName    = os.Getenv("COMMENT_FILES_TABLE_NAME")
)

// InitClient initializes the DynamoDB client with the provided config
// This should be called once during init() in main.go
func InitClient(awsConfig *aws.Config) {
	clientOnce.Do(func() {
		cfg = awsConfig
		client = dynamodb.NewFromConfig(*cfg)
	})
}
