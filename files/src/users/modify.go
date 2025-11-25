package users

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

func DeleteUserBySub(ctx context.Context, sub string) error {
	if sub == "" {
		return fmt.Errorf("DeleteUserBySub: sub is empty")
	}
	client = GetClient()
	_, err := client.AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: &userPoolId,
		Username:   &sub,
	})
	return err
}

func CreateUser(ctx context.Context, email string) (*models.User, string, error) {
	if email == "" {
		return nil, "", fmt.Errorf("CreateUser: email is empty")
	}
	client = GetClient()
	tempPassword := utils.GeneratePassword(12)
	result, err := client.AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(email),
		UserAttributes: []types.AttributeType{
			{Name: aws.String("email"), Value: aws.String(email)},
			{Name: aws.String("email_verified"), Value: aws.String("true")},
		},
		MessageAction:     types.MessageActionTypeSuppress,
		TemporaryPassword: aws.String(tempPassword),
	})
	if err != nil {
		return nil, "", err
	}
	if result == nil || result.User == nil {
		return nil, "", fmt.Errorf("CreateUser: AdminCreateUser returned nil user")
	}
	_, err = client.AdminAddUserToGroup(ctx, &cognitoidentityprovider.AdminAddUserToGroupInput{
		UserPoolId: aws.String(userPoolId),
		Username:   result.User.Username,
		GroupName:  aws.String(string(models.UserTypeUser)),
	})
	if err != nil {
		return nil, "", err
	}
	return models.UserFromCognito(*result.User, models.UserTypeUser), tempPassword, nil
}
