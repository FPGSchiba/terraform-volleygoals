package team_members

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func ListTeamMembers(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
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
	var resultItems = make([]TeamMemberListResult, 0, len(items))
	for i, item := range items {
		user := userItems[i]
		resultItems = append(resultItems, TeamMemberListResult{
			Id:                item.Id,
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
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
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
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var request UpdateTeamMemberRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	teamMember, err := db.UpdateTeamMember(ctx, teamMemberId, request.Role, request.Status)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
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
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	err := db.RemoveTeamMember(ctx, teamMemberId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
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
