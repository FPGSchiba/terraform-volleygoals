package mail

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

var (
	client     *sesv2.Client
	clientOnce sync.Once
	cfg        *aws.Config
)

func GetClient() *sesv2.Client {
	if client == nil {
		panic("❌ SESv2 client not initialized. Call InitClient first.")
	}
	return client
}
