//go:build !local

package users

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

var (
	userPoolId = os.Getenv("USER_POOL_ID")
)

// InitClient initializes the Cognito client with the provided config
// This should be called once during init() in main.go
func InitClient(awsConfig *aws.Config) {
	clientOnce.Do(func() {
		cfg = awsConfig
		client = cognitoidentityprovider.NewFromConfig(*cfg)
	})
}
