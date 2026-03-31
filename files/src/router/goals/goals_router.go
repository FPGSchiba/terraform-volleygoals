package goals

import (
	"context"
	"encoding/json"
	"math"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
	"github.com/fpgschiba/volleygoals/storage"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func CreateGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if seasonId == "" || err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	var request CreateGoalRequest
	err = json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	// Authentication and Authorization
	if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeGoals}, models.PermGoalsWrite) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	callerId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	ownerId := callerId
	if request.OwnerId != nil && utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeGoals}, models.PermGoalsWrite) {
		ownerId = *request.OwnerId
	}
	goal, err := db.CreateGoal(ctx, seasonId, ownerId, request.Type, request.Title, request.Description)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{
		"goal": goal,
	})
}

func GetGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	goalId := event.PathParameters["goalId"]
	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if seasonId == "" || goalId == "" || err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	// Authentication and Authorization
	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsRead)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"goal": goal,
	})
}

func ListGoals(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	if seasonId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	// Parse filter from query parameters
	filter, err := db.GoalFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}

	// Authorization: ensure team access via seasonId
	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if teamId == "" {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}
	if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeTeams}, models.PermTeamsRead) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	items, count, nextCursor, hasMore, err := db.ListGoals(ctx, filter)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	// Deduplicate ownerIds to minimise Cognito calls
	ownerCache := map[string]*GoalOwner{}
	for _, g := range items {
		ownerCache[g.OwnerId] = nil
	}
	for sub := range ownerCache {
		u, uerr := users.GetUserBySub(ctx, sub)
		if uerr == nil && u != nil {
			ownerCache[sub] = &GoalOwner{
				Id:                u.Id,
				Name:              u.Name,
				PreferredUsername: u.PreferredUsername,
				Picture:           u.Picture,
			}
		}
	}
	goalIds := make([]string, 0, len(items))
	for _, g := range items {
		goalIds = append(goalIds, g.Id)
	}
	progressByGoal, err := db.ListProgressEntriesByGoalIds(ctx, goalIds)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	enriched := make([]GoalWithOwner, 0, len(items))
	for _, g := range items {
		enriched = append(enriched, GoalWithOwner{
			Goal:                 g,
			Owner:                ownerCache[g.OwnerId],
			CompletionPercentage: computeCompletionPercentage(progressByGoal[g.Id]),
		})
	}

	nextToken := ""
	if nextCursor != nil {
		nextToken, err = models.EncodeCursor(nextCursor)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items":     enriched,
		"count":     count,
		"nextToken": nextToken,
		"hasMore":   hasMore,
	})
}

func UpdateGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	goalId := event.PathParameters["goalId"]
	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if seasonId == "" || goalId == "" || err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	var request UpdateGoalRequest
	err = json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.SeasonId != seasonId {
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

	userId := actorId
	updatedGoal, err := db.UpdateGoal(ctx, goalId, request.OwnerId, request.Title, request.Description, request.Status)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	if request.Status != nil {
		activity.EmitGoalStatusChanged(ctx, teamId, userId, updatedGoal.Title, *request.Status, goalId)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"goal": updatedGoal,
	})
}

func DeleteGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	goalId := event.PathParameters["goalId"]
	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if seasonId == "" || goalId == "" || err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsDelete)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	err = db.DeleteGoal(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusNoContent, utils.MsgSuccess, nil)
}

func computeCompletionPercentage(entries []*models.Progress) int {
	if len(entries) == 0 {
		return 0
	}
	var sum float64
	for _, e := range entries {
		sum += float64(e.Rating)
	}
	avg := sum / float64(len(entries))
	return int(math.Round((avg / 5.0) * 100))
}

func UploadGoalFile(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	goalId := event.PathParameters["goalId"]
	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if seasonId == "" || goalId == "" || err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	filename, ok := event.QueryStringParameters["filename"]
	if !ok || filename == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	contentType, ok := event.QueryStringParameters["contentType"]
	if !ok || contentType == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	// Authentication and Authorization
	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsWrite)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	presignedUrl, key, err := storage.GeneratePresignedUploadURLForGoalPicture(ctx, goalId, filename, contentType, utils.PresignedURLTimeout)
	if err != nil {
		return nil, err
	}
	publicUrl := storage.GetPublicFileURL(key)

	err = db.UpdateGoalPicture(ctx, goalId, publicUrl)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusOK,
		utils.MsgSuccess,
		map[string]interface{}{
			"uploadUrl": presignedUrl,
			"key":       key,
			"fileUrl":   publicUrl,
		})
}
