package db

import (
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Table Names
var teamsTableName = os.Getenv("TEAMS_TABLE_NAME")
var invitesTableName = os.Getenv("TEAM_MEMBERS_TABLE_NAME")
var teamMembersTableName = os.Getenv("INVITE_TABLE_NAME")
var teamSettingsTableName = os.Getenv("TEAM_SETTINGS_TABLE_NAME")
var seasonsTableName = os.Getenv("SEASONS_TABLE_NAME")
var goalsTableName = os.Getenv("GOALS_TABLE_NAME")
var progressReportsTableName = os.Getenv("PROGRESS_REPORTS_TABLE_NAME")
var progressTableName = os.Getenv("PROGRESS_TABLE_NAME")
var commentsTableName = os.Getenv("COMMENTS_TABLE_NAME")
var commentFilesTableName = os.Getenv("COMMENT_FILES_TABLE_NAME")

var (
	client     *dynamodb.Client
	clientOnce sync.Once
	cfg        *aws.Config
)

// InitClient initializes the DynamoDB client with the provided config
// This should be called once during init() in main.go
func InitClient(awsConfig *aws.Config) {
	clientOnce.Do(func() {
		cfg = awsConfig
		client = dynamodb.NewFromConfig(*cfg)
	})
}

// GetClient returns the initialized client
func GetClient() *dynamodb.Client {
	if client == nil {
		log.Fatal("❌ DynamoDB client not initialized. Call InitClient first.")
	}
	return client
}
