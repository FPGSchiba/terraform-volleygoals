package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

// UserType represents the Cognito group a user belongs to (ADMINS or USERS).
type UserType string

const (
	UserTypeAdmin UserType = "ADMINS"
	UserTypeUser  UserType = "USERS"
)

type UserStatus string

const (
	UserStatusConfirmed           UserStatus = "CONFIRMED"
	UserStatusUnconfirmed         UserStatus = "UNCONFIRMED"
	UserStatusExternalProvider    UserStatus = "EXTERNAL_PROVIDER"
	UserStatusResetRequired       UserStatus = "RESET_REQUIRED"
	UserStatusForceChangePassword UserStatus = "FORCE_CHANGE_PASSWORD"
	// No longer used, but still present in some user pools
	UserStatusUnknown     UserStatus = "UNKNOWN"
	UserStatusCompromised UserStatus = "COMPROMISED"
	UserStatusArchived    UserStatus = "ARCHIVED"
)

type User struct {
	Id                string     `json:"id"`
	Email             string     `json:"email"`
	Name              *string    `json:"name"`
	Picture           *string    `json:"picture"`
	PreferredUsername *string    `json:"preferredUsername"`
	Enabled           bool       `json:"enabled"`
	UserStatus        UserStatus `json:"userStatus"`
	UserType          UserType   `json:"userType"`
	UpdatedAt         *time.Time `json:"updatedAt"`
	CreatedAt         *time.Time `json:"createdAt"`
	Birthdate         *time.Time `json:"birthdate"`
}

func UserFromCognito(user types.UserType, userType UserType) *User {
	sub := ""
	email := ""
	var name *string
	var picture *string
	var preferredUsername *string
	var birthdate *time.Time

	for _, attr := range user.Attributes {
		switch *attr.Name {
		case "sub":
			sub = *attr.Value
		case "email":
			email = *attr.Value
		case "name":
			name = attr.Value
		case "picture":
			picture = attr.Value
		case "preferred_username":
			preferredUsername = attr.Value
		case "birthdate":
			if attr.Value != nil {
				if t, err := time.Parse("2006-01-02", *attr.Value); err == nil {
					birthdate = &t
				}
			}
		}
	}

	createdAt := time.Unix(user.UserCreateDate.Unix(), 0)
	updatedAt := time.Unix(user.UserLastModifiedDate.Unix(), 0)

	return &User{
		Id:                sub,
		Email:             email,
		Name:              name,
		Picture:           picture,
		PreferredUsername: preferredUsername,
		Enabled:           user.Enabled,
		UserStatus:        UserStatus(user.UserStatus),
		UserType:          userType,
		CreatedAt:         &createdAt,
		UpdatedAt:         &updatedAt,
		Birthdate:         birthdate,
	}
}

func UserFromCognitoList(users []types.UserType, userType UserType) []User {
	result := make([]User, 0, len(users))
	for _, u := range users {
		user := UserFromCognito(u, userType)
		if user != nil {
			result = append(result, *user)
		}
	}
	return result
}
