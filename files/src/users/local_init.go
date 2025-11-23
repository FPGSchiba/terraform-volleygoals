//go:build local

package users

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	log "github.com/sirupsen/logrus"
)

var (
	userPoolId = "eu-central-1_v8DJLcvpb"
)

// InitClient initializes the Cognito client for local mode. If awsConfig is
// non-nil it will be used. Otherwise this function will attempt to load the
// AWS config using an explicit profile from the environment. Environment
// variables honored (in order): LOCAL_AWS_PROFILE, AWS_PROFILE. If none are
// set the default profile behavior applies.
func InitClient(awsConfig *aws.Config) {
	clientOnce.Do(func() {
		// If an explicit aws.Config was provided, use it.
		if awsConfig != nil {
			cfg = awsConfig
			client = cognitoidentityprovider.NewFromConfig(*cfg)
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
		client = cognitoidentityprovider.NewFromConfig(*cfg)
	})
}
