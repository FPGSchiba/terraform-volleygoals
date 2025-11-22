//go:build local

package db

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	log "github.com/sirupsen/logrus"
)

// Table Names
var (
	teamsTableName           = "dev-teams"
	invitesTableName         = "dev-invites"
	teamMembersTableName     = "dev-team-members"
	teamSettingsTableName    = "dev-team-settings"
	seasonsTableName         = "dev-seasons"
	goalsTableName           = "dev-goals"
	progressReportsTableName = "dev-progress-reports"
	progressTableName        = "dev-progress"
	commentsTableName        = "dev-comments"
	commentFilesTableName    = "dev-comment-files"
)

// InitClient initializes the DynamoDB client for local mode. If awsConfig is
// non-nil it will be used. Otherwise this function will attempt to load the
// AWS config using an explicit profile from the environment. Environment
// variables honored (in order): LOCAL_AWS_PROFILE, AWS_PROFILE. If none are
// set the default profile behavior applies.
func InitClient(awsConfig *aws.Config) {
	clientOnce.Do(func() {
		// If an explicit aws.Config was provided, use it.
		if awsConfig != nil {
			cfg = awsConfig
			client = dynamodb.NewFromConfig(*cfg)
			return
		}

		// Attempt to honor a local profile first, then AWS_PROFILE
		profile := os.Getenv("LOCAL_AWS_PROFILE")
		if profile == "" {
			profile = os.Getenv("AWS_PROFILE")
		}

		var (
			c   aws.Config
			err error
		)

		if profile != "" {
			// Load config using the requested profile
			c, err = config.LoadDefaultConfig(context.Background(), config.WithSharedConfigProfile(profile))
			if err != nil {
				log.Fatalf("failed to load AWS config for profile '%s': %v", profile, err)
			}
		} else {
			// Load default config (no special profile)
			c, err = config.LoadDefaultConfig(context.Background())
			if err != nil {
				log.Fatalf("failed to load AWS config: %v", err)
			}
		}

		cfg = &c
		client = dynamodb.NewFromConfig(*cfg)
	})
}
