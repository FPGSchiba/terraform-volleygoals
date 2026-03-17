package activity

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func GetTeamActivity(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	filter, err := db.ActivityFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	filter.TeamId = teamId

	items, count, nextCursor, hasMore, err := db.ListTeamActivities(ctx, filter)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	// Members only see "all"-visibility events
	callerRole, err := utils.GetUserRoleOnTeam(ctx, event.RequestContext.Authorizer, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	isMember := !utils.IsAdmin(event.RequestContext.Authorizer) &&
		(callerRole == nil || *callerRole == models.TeamMemberRoleMember)
	if isMember {
		filtered := items[:0]
		for _, a := range items {
			if a.Visibility == models.ActivityVisibilityAll {
				filtered = append(filtered, a)
			}
		}
		items = filtered
		count = len(items)
	}

	nextToken := ""
	if nextCursor != nil {
		nextToken, err = models.EncodeCursor(nextCursor)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items":     items,
		"count":     count,
		"nextToken": nextToken,
		"hasMore":   hasMore,
	})
}

func ResolveActorInfo(u *models.User) (name string, picture string) {
	if u == nil {
		return "", ""
	}
	if u.Name != nil {
		name = *u.Name
	} else {
		name = u.Email
	}
	if u.Picture != nil {
		picture = *u.Picture
	}
	return
}

func NewActivity(teamId, actorId, actorName, actorPicture, action, description, targetType, targetId string, visibility models.ActivityVisibility) *models.Activity {
	return &models.Activity{
		Id:           models.GenerateID(),
		TeamId:       teamId,
		ActorId:      actorId,
		ActorName:    actorName,
		ActorPicture: actorPicture,
		Action:       action,
		Description:  description,
		TargetType:   targetType,
		TargetId:     targetId,
		Visibility:   visibility,
		Timestamp:    time.Now(),
	}
}

func EmitGoalStatusChanged(ctx context.Context, teamId, userId, goalTitle string, status models.GoalStatus, goalId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivity(
		teamId, userId, actorName, actorPicture,
		"goal.status_changed",
		fmt.Sprintf("Goal \"%s\" status changed to %s", goalTitle, string(status)),
		"goal", goalId,
		models.ActivityVisibilityAll,
	))
}

func EmitProgressReportCreated(ctx context.Context, teamId, userId, reportId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivity(
		teamId, userId, actorName, actorPicture,
		"progress_report.created",
		"A progress report was created",
		"progress_report", reportId,
		models.ActivityVisibilityAll,
	))
}

func EmitMemberJoined(ctx context.Context, teamId, userId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivity(
		teamId, userId, actorName, actorPicture,
		"member.joined",
		fmt.Sprintf("%s joined the team", actorName),
		"team_member", "",
		models.ActivityVisibilityAll,
	))
}

func EmitMemberRoleChanged(ctx context.Context, teamId, userId string, role models.TeamMemberRole, memberId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivity(
		teamId, userId, actorName, actorPicture,
		"member.role_changed",
		fmt.Sprintf("A member's role was changed to %s", string(role)),
		"team_member", memberId,
		models.ActivityVisibilityAdminTrainer,
	))
}

func EmitMemberRemoved(ctx context.Context, teamId, userId, memberId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivity(
		teamId, userId, actorName, actorPicture,
		"member.removed",
		"A member was removed from the team",
		"team_member", memberId,
		models.ActivityVisibilityAdminTrainer,
	))
}

func EmitTeamSettingsUpdated(ctx context.Context, teamId, userId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivity(
		teamId, userId, actorName, actorPicture,
		"team_settings.updated",
		"Team settings were updated",
		"team_settings", teamId,
		models.ActivityVisibilityAdminTrainer,
	))
}
