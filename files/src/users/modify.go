package users

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

func DeleteUserBySub(ctx context.Context, sub string) error {
	client = GetClient()
	_, err := client.AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: &userPoolId,
		Username:   &sub,
	})
	return err
}
