package db

import (
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var (
	client     *dynamodb.Client
	clientOnce sync.Once
	cfg        *aws.Config
)

// GetClient returns the initialized client
func GetClient() *dynamodb.Client {
	if client == nil {
		log.Fatal("❌ DynamoDB client not initialized. Call InitClient first.")
	}
	return client
}
