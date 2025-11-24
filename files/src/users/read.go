package users

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/fpgschiba/volleygoals/models"
)

// ErrUserNotFound is returned when a requested user does not exist in Cognito
var ErrUserNotFound = errors.New("user not found")

type ListUserResult struct {
	PaginationToken *string
	Users           []models.User
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return aws.String(s)
}

func buildListUsersInGroupInput(f *UserFilter) *cognitoidentityprovider.ListUsersInGroupInput {
	return &cognitoidentityprovider.ListUsersInGroupInput{
		UserPoolId: aws.String(userPoolId),
		GroupName:  aws.String(f.GroupName),
		Limit:      aws.Int32(f.limitOrDefault(25)),
		NextToken:  strPtrOrNil(f.PaginationToken),
	}
}

func buildListUsersInput(f *UserFilter) *cognitoidentityprovider.ListUsersInput {
	return &cognitoidentityprovider.ListUsersInput{
		UserPoolId:      aws.String(userPoolId),
		Filter:          strPtrOrNil(f.Filter),
		Limit:           aws.Int32(f.limitOrDefault(25)),
		PaginationToken: strPtrOrNil(f.PaginationToken),
	}
}

// ListAdminUsers lists users in the admin group using a UserFilter.
func ListAdminUsers(ctx context.Context, filter *UserFilter) (*ListUserResult, error) {
	if filter == nil {
		filter = &UserFilter{}
	}
	filter.GroupName = string(models.UserTypeAdmin)

	client = GetClient()
	in := buildListUsersInGroupInput(filter)
	result, err := client.ListUsersInGroup(ctx, in)
	if err != nil {
		return nil, err
	}
	return &ListUserResult{
		PaginationToken: result.NextToken,
		Users:           models.UserFromCognitoList(result.Users, models.UserTypeAdmin),
	}, nil
}

// ListUsers lists members of the regular user group using a UserFilter.
func ListUsers(ctx context.Context, filter *UserFilter) (*ListUserResult, error) {
	if filter == nil {
		filter = &UserFilter{}
	}
	filter.GroupName = string(models.UserTypeUser)

	client = GetClient()
	in := buildListUsersInGroupInput(filter)
	result, err := client.ListUsersInGroup(ctx, in)
	if err != nil {
		return nil, err
	}
	return &ListUserResult{
		PaginationToken: result.NextToken,
		Users:           models.UserFromCognitoList(result.Users, models.UserTypeUser),
	}, nil
}

// ListAllUsers lists users across the pool (supports Cognito filter and pagination).
func ListAllUsers(ctx context.Context, filter *UserFilter) (*ListUserResult, error) {
	if filter == nil {
		filter = &UserFilter{}
	}

	client = GetClient()
	in := buildListUsersInput(filter)
	result, err := client.ListUsers(ctx, in)
	if err != nil {
		return nil, err
	}
	var users []models.User
	for _, u := range result.Users {
		out, err := client.AdminListGroupsForUser(ctx, &cognitoidentityprovider.AdminListGroupsForUserInput{
			UserPoolId: aws.String(userPoolId),
			Username:   u.Username,
			Limit:      aws.Int32(1), // Max 1 group needed to determine type
		})
		if err != nil {
			// If the user doesn't exist, map to ErrUserNotFound so callers can distinguish
			var notFound *types.UserNotFoundException
			if errors.As(err, &notFound) {
				return nil, ErrUserNotFound
			}
			return nil, err
		}
		var userType models.UserType = models.UserTypeUser
		if len(out.Groups) > 0 && out.Groups[0].GroupName != nil && *out.Groups[0].GroupName == string(models.UserTypeAdmin) {
			userType = models.UserTypeAdmin
		}
		users = append(users, *models.UserFromCognito(u, userType))
	}
	return &ListUserResult{
		PaginationToken: result.PaginationToken,
		Users:           users,
	}, nil
}

func typesToUserType(output *cognitoidentityprovider.AdminGetUserOutput) types.UserType {
	return types.UserType{
		Username:             output.Username,
		Attributes:           output.UserAttributes,
		UserCreateDate:       output.UserCreateDate,
		UserLastModifiedDate: output.UserLastModifiedDate,
		Enabled:              output.Enabled,
		UserStatus:           output.UserStatus,
	}
}

func GetUserBySub(ctx context.Context, sub string) (*models.User, error) {
	client = GetClient()
	result, err := client.AdminGetUser(ctx, &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(sub),
	})
	if err != nil {
		var notFound *types.UserNotFoundException
		if errors.As(err, &notFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	out, err := client.AdminListGroupsForUser(ctx, &cognitoidentityprovider.AdminListGroupsForUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   aws.String(sub),
		Limit:      aws.Int32(1), // Max 1 group needed to determine type
	})
	if err != nil {
		var notFound *types.UserNotFoundException
		if errors.As(err, &notFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	var userType = models.UserTypeUser
	if len(out.Groups) > 0 && out.Groups[0].GroupName != nil && *out.Groups[0].GroupName == string(models.UserTypeAdmin) {
		userType = models.UserTypeAdmin
	}
	user := models.UserFromCognito(typesToUserType(result), userType)
	return user, nil
}

func GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	client = GetClient()
	result, err := client.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(userPoolId),
		Filter:     aws.String("email = \"" + email + "\""),
		Limit:      aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Users) == 0 {
		return nil, nil
	}
	out, err := client.AdminListGroupsForUser(ctx, &cognitoidentityprovider.AdminListGroupsForUserInput{
		UserPoolId: aws.String(userPoolId),
		Username:   result.Users[0].Username,
		Limit:      aws.Int32(1), // Max 1 group needed to determine type
	})
	if err != nil {
		var notFound *types.UserNotFoundException
		if errors.As(err, &notFound) {
			return nil, nil
		}
		return nil, err
	}
	var userType = models.UserTypeUser
	if len(out.Groups) > 0 && out.Groups[0].GroupName != nil && *out.Groups[0].GroupName == string(models.UserTypeAdmin) {
		userType = models.UserTypeAdmin
	}
	user := models.UserFromCognito(result.Users[0], userType)
	return user, nil
}
