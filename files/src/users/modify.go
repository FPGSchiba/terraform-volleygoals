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

func UpdateUserAttributes(ctx context.Context, sub string, attributes map[string]string) error {
	if sub == "" {
		return fmt.Errorf("UpdateUserAttributes: sub is empty")
	}
	if len(attributes) == 0 {
		return fmt.Errorf("UpdateUserAttributes: attributes is empty")
	}
	client = GetClient()
	var cognitoAttributes []types.AttributeType
	for k, v := range attributes {
		cognitoAttributes = append(cognitoAttributes, types.AttributeType{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}
	_, err := client.AdminUpdateUserAttributes(ctx, &cognitoidentityprovider.AdminUpdateUserAttributesInput{
		UserPoolId:     aws.String(userPoolId),
		Username:       aws.String(sub),
		UserAttributes: cognitoAttributes,
	})
	return err
}

func UpdateUserType(ctx context.Context, sub string, newType models.UserType) error {
	if sub == "" {
		return fmt.Errorf("UpdateUserGroup: sub is empty")
	}
	client = GetClient()
	// First, get the current groups of the user
	groupsResult, err := client.AdminListGroupsForUser(ctx, &cognitoidentityprovider.AdminListGroupsForUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(sub),
	})
	if err != nil {
		return err
	}
	// Remove user from all current groups
	for _, group := range groupsResult.Groups {
		_, err := client.AdminRemoveUserFromGroup(ctx, &cognitoidentityprovider.AdminRemoveUserFromGroupInput{
			UserPoolId: aws.String(userPoolId),
			Username:   aws.String(sub),
			GroupName:  group.GroupName,
		})
		if err != nil {
			return err
		}
	}
	// Add user to the new group
	_, err = client.AdminAddUserToGroup(ctx, &cognitoidentityprovider.AdminAddUserToGroupInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(sub),
		GroupName:  aws.String(string(newType)),
	})
	return err
}

func DisableUser(ctx context.Context, sub string) error {
	if sub == "" {
		return fmt.Errorf("DisableUser: sub is empty")
	}
	client = GetClient()
	_, err := client.AdminDisableUser(ctx, &cognitoidentityprovider.AdminDisableUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(sub),
	})
	return err
}

func EnableUser(ctx context.Context, sub string) error {
	if sub == "" {
		return fmt.Errorf("EnableUser: sub is empty")
	}
	client = GetClient()
	_, err := client.AdminEnableUser(ctx, &cognitoidentityprovider.AdminEnableUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(sub),
	})
	return err
}
