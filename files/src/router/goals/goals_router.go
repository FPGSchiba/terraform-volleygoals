package goals

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/storage"
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
	if request.Type == models.GoalTypeTeam {
		// Verify if the requester is a team admin or trainer
		if !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	} else {
		// For individual goals, verify the user has team access
		if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	ownerId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
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

	// Authentication and Authorization
	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
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
	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	items, count, nextCursor, hasMore, err := db.ListGoals(ctx, filter)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	nextToken := ""
	if nextCursor != nil {
		nextToken, err = models.EncodeCursor(nextCursor)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
	}

	resp := models.PaginationResponse{
		Items:     items,
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

	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	// Only the owner or team admin/trainer can update the goal
	if goal.OwnerId != userId && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	updatedGoal, err := db.UpdateGoal(ctx, goalId, request.OwnerId, request.Title, request.Description, request.Status)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
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

	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	// Only the owner or team admin/trainer can delete the goal
	if goal.OwnerId != userId && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	err = db.DeleteGoal(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusNoContent, utils.MsgSuccess, nil)
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

	// Authentication and Authorization
	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	// Only the owner or team admin/trainer can upload files to the goal
	if goal.OwnerId != userId && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
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
