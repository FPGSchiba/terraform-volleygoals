package utils

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
)

type UserType string

type AccessLevel string

const (
	UserTypeAdmin UserType = "ADMINS"
	UserTypeUser  UserType = "USERS"
)

const (
	AccessLevelRead  AccessLevel = "READ"
	AccessLevelWrite AccessLevel = "WRITE"
)

func IsAdmin(authorizer map[string]interface{}) bool {
	return hasUserType(authorizer, []UserType{UserTypeAdmin})
}

func IsUser(authorizer map[string]interface{}) bool {
	return hasUserType(authorizer, []UserType{UserTypeUser})
}

func hasUserType(authorizer map[string]interface{}, allowedRoles []UserType) bool {
	if authorizer == nil {
		return false
	}
	claims, ok := authorizer["claims"]
	if !ok {
		return false
	}
	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		return false
	}
	groups, ok := claimsMap["cognito:groups"]
	if !ok {
		return false
	}

	var groupsSlice []string
	switch g := groups.(type) {
	case []string:
		groupsSlice = g
	case []interface{}:
		for _, gi := range g {
			if s, ok := gi.(string); ok {
				groupsSlice = append(groupsSlice, s)
			}
		}
	case string:
		groupsSlice = []string{g}
	default:
		return false
	}

	if len(groupsSlice) == 0 {
		return false
	}

	groupSet := make(map[string]struct{}, len(groupsSlice))
	for _, g := range groupsSlice {
		groupSet[g] = struct{}{}
	}

	for _, role := range allowedRoles {
		if _, exists := groupSet[string(role)]; exists {
			return true
		}
	}

	return false
}

func GetCognitoUsername(authorizer map[string]interface{}) string {
	if authorizer == nil {
		return ""
	}
	claims, ok := authorizer["claims"]
	if !ok {
		return ""
	}
	claimsMap, ok := claims.(map[string]interface{})
	if !ok {
		return ""
	}
	sub, ok := claimsMap["cognito:username"]
	if !ok {
		return ""
	}
	subStr, ok := sub.(string)
	if !ok {
		return ""
	}
	return subStr
}

func HasOneRoleOnTeam(ctx context.Context, authorizer map[string]interface{}, teamId string, requiredRole []models.TeamMemberRole) bool {
	if !IsUser(authorizer) {
		return false
	}
	userID := GetCognitoUsername(authorizer)
	for _, role := range requiredRole {
		ok, err := db.HasRoleOnTeam(ctx, userID, teamId, role)
		if err != nil {
			return false
		}
		if ok {
			return true
		}
	}
	return false
}
