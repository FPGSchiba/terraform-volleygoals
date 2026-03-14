package comments

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

func CreateComment(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var request CreateCommentRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	if request.TargetId == "" || string(request.CommentType) == "" || request.Content == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	// Resolve teamId from target
	teamId, err := resolveTeamIdFromTarget(ctx, request.CommentType, request.TargetId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if teamId == "" {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	// Check TeamSettings for Goal comment types
	if request.CommentType == models.CommentTypeGoal {
		goal, err := db.GetGoalById(ctx, request.TargetId)
		if err != nil || goal == nil {
			return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
		}
		settings, err := db.GetTeamSettingsByTeamID(ctx, teamId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
		}
		if settings != nil {
			if goal.GoalType == models.GoalTypeTeam && !settings.AllowTeamGoalComments {
				return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorCommentsDisabled, nil)
			}
			if goal.GoalType == models.GoalTypeIndividual && !settings.AllowIndividualGoalComments {
				return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorCommentsDisabled, nil)
			}
		}
	}

	authorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	comment, err := db.CreateComment(ctx, authorId, string(request.CommentType), request.TargetId, request.Content)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{
		"comment": comment,
	})
}

func GetComment(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	commentId := event.PathParameters["commentId"]
	if commentId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	comment, err := db.GetCommentById(ctx, commentId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if comment == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorCommentNotFound, nil)
	}

	teamId, err := resolveTeamIdFromTarget(ctx, comment.CommentType, comment.TargetId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"comment": comment,
	})
}

func ListComments(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	filter, err := db.CommentFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}

	teamId, err := resolveTeamIdFromTarget(ctx, models.CommentType(filter.CommentType), filter.TargetId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if teamId == "" {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	items, count, nextCursor, hasMore, err := db.ListComments(ctx, filter)
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

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items":     items,
		"count":     count,
		"nextToken": nextToken,
		"hasMore":   hasMore,
	})
}

func UpdateComment(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	commentId := event.PathParameters["commentId"]
	if commentId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	comment, err := db.GetCommentById(ctx, commentId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if comment == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorCommentNotFound, nil)
	}

	teamId, err := resolveTeamIdFromTarget(ctx, comment.CommentType, comment.TargetId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if comment.AuthorId != userId && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	var request UpdateCommentRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	updatedComment, err := db.UpdateComment(ctx, commentId, request.Content)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"comment": updatedComment,
	})
}

func DeleteComment(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	commentId := event.PathParameters["commentId"]
	if commentId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	comment, err := db.GetCommentById(ctx, commentId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if comment == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorCommentNotFound, nil)
	}

	teamId, err := resolveTeamIdFromTarget(ctx, comment.CommentType, comment.TargetId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if comment.AuthorId != userId && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	if err := db.DeleteComment(ctx, commentId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusNoContent, utils.MsgSuccess, nil)
}

func UploadCommentFile(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	commentId := event.PathParameters["commentId"]
	if commentId == "" {
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

	comment, err := db.GetCommentById(ctx, commentId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if comment == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorCommentNotFound, nil)
	}

	teamId, err := resolveTeamIdFromTarget(ctx, comment.CommentType, comment.TargetId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if comment.AuthorId != userId && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	presignedUrl, key, err := storage.GeneratePresignedUploadURLForCommentFile(ctx, commentId, filename, contentType, utils.PresignedURLTimeout)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	cf, err := db.CreateCommentFile(ctx, commentId, key)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"uploadUrl":   presignedUrl,
		"key":         key,
		"commentFile": cf,
	})
}

// resolveTeamIdFromTarget looks up the teamId for a comment target (Goal or ProgressReport).
func resolveTeamIdFromTarget(ctx context.Context, commentType models.CommentType, targetId string) (string, error) {
	switch commentType {
	case models.CommentTypeGoal:
		goal, err := db.GetGoalById(ctx, targetId)
		if err != nil {
			return "", err
		}
		if goal == nil {
			return "", nil
		}
		return db.GetTeamIdBySeasonId(ctx, goal.SeasonId)
	case models.CommentTypeProgressReport:
		report, err := db.GetProgressReportById(ctx, targetId)
		if err != nil {
			return "", err
		}
		if report == nil {
			return "", nil
		}
		return db.GetTeamIdBySeasonId(ctx, report.SeasonId)
	default:
		return "", nil
	}
}
