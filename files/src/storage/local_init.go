//go:build local

package storage

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
)

var (
	bucketName = "dev-volleygoals-20251129110008477200000001"
	cdnBaseURL = "https://cdn.volleygoals-test.schiba-apps.net"
)

// InitClient initializes the S3 client for local mode. If awsConfig is
// non-nil it will be used. Otherwise this function will attempt to load the
// AWS config using an explicit profile from the environment. Environment
// variables honored (in order): LOCAL_AWS_PROFILE, AWS_PROFILE. If none are
// set the default profile behavior applies.
func InitClient(awsConfig *aws.Config) {
	clientOnce.Do(func() {
		// If an explicit aws.Config was provided, use it.
		if awsConfig != nil {
			cfg = awsConfig
			client = s3.NewFromConfig(*cfg)
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
		client = s3.NewFromConfig(*cfg)
		initPresignClient(client)
	})
}

func initPresignClient(client *s3.Client) {
	presignClientOnce.Do(func() {
		presignClient = s3.NewPresignClient(client)
	})
}
