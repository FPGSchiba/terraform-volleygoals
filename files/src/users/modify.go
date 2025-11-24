package users

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

func DeleteUserBySub(ctx context.Context, sub string) error {
	client = GetClient()
	_, err := client.AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: &userPoolId,
		Username:   &sub,
	})
	return err
}

func CreateUser(ctx context.Context, email string) (*models.User, error) {
	client = GetClient()
	result, err := client.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(models.GenerateID()),
		UserAttributes: []types.AttributeType{
			{Name: aws.String("email"), Value: aws.String(email)},
			{Name: aws.String("email_verified"), Value: aws.String("true")},
		},
		MessageAction:     types.MessageActionTypeSuppress,
		TemporaryPassword: aws.String(utils.GeneratePassword(12)),
	})
	if err != nil {
		return nil, err
	}
	_, err = client.AdminAddUserToGroup(ctx, &cognitoidentityprovider.AdminAddUserToGroupInput{
		UserPoolId: aws.String(userPoolId),
		Username:   result.User.Username,
		GroupName:  aws.String(string(models.UserTypeUser)),
	})
	if err != nil {
		return nil, err
	}
	return models.UserFromCognito(*result.User, models.UserTypeUser), nil
}
