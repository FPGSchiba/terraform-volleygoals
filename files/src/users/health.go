package users

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

// CheckHealth verifies Cognito connectivity by describing the user pool.
func CheckHealth(ctx context.Context) error {
	client = GetClient()
	_, err := client.DescribeUserPool(ctx, &cognitoidentityprovider.DescribeUserPoolInput{
		UserPoolId: aws.String(userPoolId),
	})
	return err
}
