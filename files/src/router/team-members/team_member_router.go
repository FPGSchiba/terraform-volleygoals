package team_members

import (
	"context"
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
	// TODO: implement adding a team member
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func UpdateTeamMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement updating a team member
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func RemoveTeamMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement removing a team member
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func LeaveTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement leaving a team
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}
