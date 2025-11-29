//go:build !local

package storage

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	bucketName = os.Getenv("S3_BUCKET_NAME")
	cdnBaseURL = os.Getenv("CDN_BASE_URL")
)

// InitClient initializes the S3 client with the provided config
// This should be called once during init() in main.go
func InitClient(awsConfig *aws.Config) {
	clientOnce.Do(func() {
		cfg = awsConfig
		client = s3.NewFromConfig(*cfg)
		initPresignClient(client)
	})
}

// InitPresignClient initializes the S3 client with the provided config
// This should be called once during init() in main.go
func initPresignClient(client *s3.Client) {
	presignClientOnce.Do(func() {
		presignClient = s3.NewPresignClient(client)
	})
}
