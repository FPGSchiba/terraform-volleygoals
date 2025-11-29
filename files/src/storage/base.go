package storage

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
)

var (
	client            *s3.Client
	presignClient     *s3.PresignClient
	clientOnce        sync.Once
	presignClientOnce sync.Once
	cfg               *aws.Config
)

// GetClient returns the initialized client
func GetClient() *s3.Client {
	if client == nil {
		log.Fatal("❌ DynamoDB client not initialized. Call InitClient first.")
	}
	return client
}

// GetPresignClient returns the initialized presign client
func GetPresignClient() *s3.PresignClient {
	if presignClient == nil {
		log.Fatal("❌ S3 Presign client not initialized. Call InitClient first.")
	}
	return presignClient
}
