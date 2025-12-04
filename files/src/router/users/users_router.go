package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func ListUsers(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	filter, err := users.UserFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	var result *users.ListUserResult
	switch strings.ToLower(strings.TrimSpace(filter.GroupName)) {
	case "admin":
	case "admins":
		result, err = users.ListAdminUsers(ctx, filter)
	case "user":
	case "users":
		result, err = users.ListUsers(ctx, filter)
	default:
		result, err = users.ListAllUsers(ctx, filter)
	}
	if err != nil || result == nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"users":           result.Users,
		"paginationToken": result.PaginationToken,
	})
}

func GetUser(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	userSub, ok := event.PathParameters["userSub"]
	if !ok || userSub == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	user, err := users.GetUserBySub(ctx, userSub)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, err)
		}
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	memberships, err := db.GetMembershipsByUserID(ctx, user.Id)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"user":        user,
		"memberships": memberships,
	})
}

func DeleteUser(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	userSub, ok := event.PathParameters["userSub"]
	if !ok || userSub == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	err := users.DeleteUserBySub(ctx, userSub)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	err = db.DeleteTeamMembershipsByUserID(ctx, userSub)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}

func UpdateUser(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	userSub, ok := event.PathParameters["userSub"]
	if !ok || userSub == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	var request UpdateUserRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}

	user, err := applyUserTypeUpdate(ctx, userSub, request.UserType)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if request.UserType != nil && user == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	if err := applyEnabledUpdate(ctx, userSub, request.Enabled); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	
	user, err = users.GetUserBySub(ctx, userSub)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"user": user,
	})
}

func applyUserTypeUpdate(ctx context.Context, userSub string, userType *models.UserType) (*models.User, error) {
	if userType == nil {
		return nil, nil
	}
	user, err := users.GetUserBySub(ctx, userSub)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	if *userType == models.UserTypeAdmin {
		if err := db.RemoveAllTeamMembersForUser(ctx, user.Id); err != nil {
			return nil, err
		}
	}
	if err := users.UpdateUserType(ctx, userSub, *userType); err != nil {
		return nil, err
	}
	return user, nil
}

func applyEnabledUpdate(ctx context.Context, userSub string, enabled *bool) error {
	if enabled == nil {
		return nil
	}
	if *enabled {
		return users.EnableUser(ctx, userSub)
	}
	return users.DisableUser(ctx, userSub)
}
