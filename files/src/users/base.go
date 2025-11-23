package users

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	log "github.com/sirupsen/logrus"
)

var (
	client     *cognitoidentityprovider.Client
	clientOnce sync.Once
	cfg        *aws.Config
)

// GetClient returns the initialized client
func GetClient() *cognitoidentityprovider.Client {
	if client == nil {
		log.Fatal("❌ DynamoDB client not initialized. Call InitClient first.")
	}
	return client
}
