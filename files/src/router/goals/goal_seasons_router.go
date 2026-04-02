package goals

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

func TagGoalToSeason(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	seasonId := event.PathParameters["seasonId"]
	if teamId == "" || goalId == "" || seasonId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsWrite)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	gs, err := db.TagGoalToSeason(ctx, goalId, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusConflict, "goal already tagged to this season", nil)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{
		"goalSeason": gs,
	})
}

func UntagGoalFromSeason(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	seasonId := event.PathParameters["seasonId"]
	if teamId == "" || goalId == "" || seasonId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsWrite)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	if err := db.UntagGoalFromSeason(ctx, goalId, seasonId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	return utils.SuccessResponse(http.StatusNoContent, utils.MsgSuccess, nil)
}

func ListGoalSeasons(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	if teamId == "" || goalId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsRead)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	items, err := db.ListSeasonsByGoalId(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}
