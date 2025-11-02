package db

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

var client *dynamodb.Client

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

func GetClient() *dynamodb.Client {
	if client == nil {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Fatalf("unable to load AWS-SDK config, %v", err)
		}

		otelaws.AppendMiddlewares(&cfg.APIOptions)
		client = dynamodb.NewFromConfig(cfg)
		return client
	} else {
		return client
	}
}
