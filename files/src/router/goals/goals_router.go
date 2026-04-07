package goals

import (
	"context"
	"encoding/json"
	"math"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/db/instrumented"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
	"github.com/fpgschiba/volleygoals/storage"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func CreateGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	var request CreateGoalRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	callerId := utils.GetCognitoUsername(event.RequestContext.Authorizer)

	rt := models.ResourceTypeTeamGoals
	rp := models.PermTeamGoalsWrite
	// For individual goals the caller will own the new resource, so ownership policy
	// applies (any team member may create their own individual goal). For team goals,
	// creation is role-gated only — members must not bypass via ownership.
	resource := models.Resource{Type: rt}
	if request.Type == models.GoalTypeIndividual {
		rt = models.ResourceTypeIndividualGoals
		rp = models.PermIndividualGoalsWrite
		resource = models.Resource{Type: rt, OwnedBy: callerId}
	}

	allowed, err := utils.CheckPermission(ctx, callerId, teamId, resource, rp)
	if err != nil || !allowed {
		if !utils.IsAdmin(event.RequestContext.Authorizer) {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	goal, err := instrumented.CreateGoal(ctx, teamId, callerId, request.Type, request.Title, request.Description)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{
		"goal": goal,
	})
}

func GetGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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
		rt := goal.GetResourceType()
		rp := models.PermTeamGoalsRead
		if goal.GoalType == models.GoalTypeIndividual {
			rp = models.PermIndividualGoalsRead
		}
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: rt, OwnedBy: goal.OwnerId},
			rp)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"goal": goal,
	})
}

func ListGoals(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	filter, err := db.GoalFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	filter.TeamId = teamId

	// Optional season filter: if ?seasonId= is provided, restrict to goals tagged to that season.
	if sid, ok := event.QueryStringParameters["seasonId"]; ok && sid != "" {
		goalIds, err := db.ListGoalIdsBySeasonId(ctx, sid)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		filter.GoalIds = goalIds
		if len(goalIds) == 0 {
			return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
				"items": []interface{}{}, "count": 0, "nextToken": "", "hasMore": false,
			})
		}
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	// We no longer strictly bound `filter.OwnerId = actorId` because that breaks
	// visibility of team goals and managed individual goals. We let DB fetch relevant goals
	// by team / season, and then filter out unauthorized ones post-fetch to respect Epic 1.3 rules.

	items, count, nextCursor, hasMore, err := db.ListGoals(ctx, filter)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	var allowedItems []*models.Goal
	for _, g := range items {
		if utils.IsAdmin(event.RequestContext.Authorizer) {
			allowedItems = append(allowedItems, g)
			continue
		}
		rt := g.GetResourceType()
		rp := models.PermTeamGoalsRead
		if g.GoalType == models.GoalTypeIndividual {
			rp = models.PermIndividualGoalsRead
		}
		allowed, _ := utils.CheckPermission(ctx, actorId, teamId, models.Resource{Type: rt, OwnedBy: g.OwnerId}, rp)
		if allowed {
			allowedItems = append(allowedItems, g)
		}
	}
	items = allowedItems
	count = len(items)

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
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	if teamId == "" || goalId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	var request UpdateGoalRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
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
		rt := goal.GetResourceType()
		rp := models.PermTeamGoalsWrite
		if goal.GoalType == models.GoalTypeIndividual {
			rp = models.PermIndividualGoalsWrite
		}
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: rt, OwnedBy: goal.OwnerId},
			rp)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	updatedGoal, err := db.UpdateGoal(ctx, goalId, request.OwnerId, request.Title, request.Description, request.Status)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	if request.Status != nil {
		activity.EmitGoalStatusChanged(ctx, teamId, actorId, updatedGoal.Title, *request.Status, goalId, updatedGoal.OwnerId)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"goal": updatedGoal,
	})
}

func DeleteGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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
		rt := goal.GetResourceType()
		rp := models.PermTeamGoalsDelete
		if goal.GoalType == models.GoalTypeIndividual {
			rp = models.PermIndividualGoalsDelete
		}
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: rt, OwnedBy: goal.OwnerId},
			rp)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	if err := instrumented.DeleteGoal(ctx, teamId, actorId, goalId, goal.Title, goal.OwnerId); err != nil {
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
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	if teamId == "" || goalId == "" {
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
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		rt := goal.GetResourceType()
		rp := models.PermTeamGoalsWrite
		if goal.GoalType == models.GoalTypeIndividual {
			rp = models.PermIndividualGoalsWrite
		}
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: rt, OwnedBy: goal.OwnerId},
			rp)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	presignedUrl, key, err := storage.GeneratePresignedUploadURLForGoalPicture(ctx, goalId, filename, contentType, utils.PresignedURLTimeout)
	if err != nil {
		return nil, err
	}
	publicUrl := storage.GetPublicFileURL(key)

	if err := db.UpdateGoalPicture(ctx, goalId, publicUrl); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"uploadUrl": presignedUrl,
		"key":       key,
		"fileUrl":   publicUrl,
	})
}
