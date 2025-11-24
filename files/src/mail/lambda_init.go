//go:build !local

package mail

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

var (
	EmailSender          = os.Getenv("EMAIL_SENDER")
	TenantName           = os.Getenv("TENANT_NAME")
	ConfigurationSetName = os.Getenv("CONFIGURATION_SET_NAME")
	FrontendBaseUrl      = os.Getenv("FRONTEND_BASE_URL")
	InviteTemplateArn    = os.Getenv("INVITE_TEMPLATE_ARN")
)

// InitClient initializes the DynamoDB client with the provided config
// This should be called once during init() in main.go
func InitClient(awsConfig *aws.Config) {
	clientOnce.Do(func() {
		cfg = awsConfig
		client = sesv2.NewFromConfig(*cfg)
	})
}
