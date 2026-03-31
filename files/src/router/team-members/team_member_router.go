package team_members

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func ListTeamMembers(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeMembers}, models.PermMembersRead) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	callerRole, err := utils.GetUserRoleOnTeam(ctx, event.RequestContext.Authorizer, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	isMember := callerRole != nil && *callerRole == models.TeamMemberRoleMember
	filter, err := db.TeamMemberFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	items, count, nextCursor, hasMore, err := db.ListTeamMembers(ctx, teamId, filter)
	if err != nil {
		return nil, err
	}
	nextToken := ""
	if nextCursor != nil {
		nextToken, err = models.EncodeCursor(nextCursor)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
	}
	userItems, err := users.GetUsersByTeamMembers(ctx, items)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	// Apply in-memory name/email filtering (name/email are not stored in DynamoDB).
	// When either filter is active, cursor pagination is not accurate; all matching
	// results are returned regardless of the cursor/limit settings.
	if filter.NameContains != "" || filter.EmailContains != "" {
		nameLower := strings.ToLower(filter.NameContains)
		emailLower := strings.ToLower(filter.EmailContains)
		filteredItems := items[:0]
		filteredUsers := userItems[:0]
		for i, u := range userItems {
			nameVal := ""
			if u.Name != nil {
				nameVal = strings.ToLower(*u.Name)
			}
			emailVal := strings.ToLower(u.Email)
			if nameLower != "" && !strings.Contains(nameVal, nameLower) {
				continue
			}
			if emailLower != "" && !strings.Contains(emailVal, emailLower) {
				continue
			}
			filteredItems = append(filteredItems, items[i])
			filteredUsers = append(filteredUsers, u)
		}
		items = filteredItems
		userItems = filteredUsers
		count = len(items)
		nextToken = ""
		hasMore = false
	}

	if isMember {
		publicItems := make([]TeamMemberPublicResult, 0, len(items))
		for i, item := range items {
			user := userItems[i]
			publicItems = append(publicItems, TeamMemberPublicResult{
				Id:                item.Id,
				UserId:            item.UserId,
				Name:              user.Name,
				PreferredUsername: user.PreferredUsername,
				Picture:           user.Picture,
				Email:             user.Email,
			})
		}
		return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
			"items":     publicItems,
			"count":     count,
			"nextToken": nextToken,
			"hasMore":   hasMore,
		})
	}
	var resultItems = make([]TeamMemberListResult, 0, len(items))
	for i, item := range items {
		user := userItems[i]
		resultItems = append(resultItems, TeamMemberListResult{
			Id:                item.Id,
			UserId:            item.UserId,
			Role:              item.Role,
			Status:            item.Status,
			UserStatus:        user.UserStatus,
			Name:              user.Name,
			Email:             user.Email,
			Picture:           user.Picture,
			PreferredUsername: user.PreferredUsername,
			Birthdate:         user.Birthdate,
			JoinedAt:          item.JoinedAt,
		})
	}
	resp := models.PaginationResponse{
		Items:     resultItems,
		Count:     count,
		NextToken: nextToken,
		HasMore:   hasMore,
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items":     resp.Items,
		"count":     resp.Count,
		"nextToken": resp.NextToken,
		"hasMore":   resp.HasMore,
	})
}

func AddTeamMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeMembers}, models.PermMembersWrite) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var request AddTeamMemberRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	user, err := users.GetUserBySub(ctx, request.UserId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if user == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorUserNotFound, nil)
	}
	teamMember, err := db.AddTeamMember(ctx, teamId, request.UserId, request.Role)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	activity.EmitMemberJoined(ctx, teamId, request.UserId)

	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{
		"teamMember": teamMember,
	})
}

func UpdateTeamMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	teamMemberId, ok := event.PathParameters["memberId"]
	if !ok || teamMemberId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeMembers}, models.PermMembersWrite) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var request UpdateTeamMemberRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	if request.Role != nil && *request.Role == models.TeamMemberRoleAdmin {
		if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeMembers}, models.PermMembersDelete) {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	teamMember, err := db.UpdateTeamMember(ctx, teamMemberId, request.Role, request.Status)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	if request.Role != nil {
		userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
		activity.EmitMemberRoleChanged(ctx, teamId, userId, *request.Role, teamMemberId)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"teamMember": teamMember,
	})
}

func RemoveTeamMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	teamMemberId, ok := event.PathParameters["memberId"]
	if !ok || teamMemberId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeMembers}, models.PermMembersWrite) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	err := db.RemoveTeamMember(ctx, teamMemberId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	activity.EmitMemberRemoved(ctx, teamId, userId, teamMemberId)

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}

func LeaveTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if userId == "" {
		return utils.ErrorResponse(http.StatusUnauthorized, utils.MsgErrorUnauthorized, nil)
	}
	userRole, err := utils.GetUserRoleOnTeam(ctx, event.RequestContext.Authorizer, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if userRole == nil {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	if *userRole == models.TeamMemberRoleAdmin || *userRole == models.TeamMemberRoleTrainer {
		// Only able to leave if another Trainer or Admin exists
		hasOther, err := db.HasOtherAdminOrTrainer(ctx, teamId, userId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if !hasOther {
			return utils.ErrorResponse(http.StatusNotAcceptable, utils.MsgErrorMemberCannotLeave, nil)
		}
	}
	err = db.LeaveTeam(ctx, teamId, userId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}
