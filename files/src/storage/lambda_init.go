//go:build !local

package storage

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// InitClient initializes the S3 client with the provided config
// This should be called once during init() in main.go
func InitClient(awsConfig *aws.Config) {
	clientOnce.Do(func() {
		cfg = awsConfig
		client = s3.NewFromConfig(*cfg)
	})
}
