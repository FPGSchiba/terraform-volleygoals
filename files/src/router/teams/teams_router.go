package teams

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

func UpdateTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	var request UpdateTeamRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	team, err := db.GetTeamById(ctx, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if team == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTeamNotFound, nil)
	}
	if request.Name != nil {
		team.Name = *request.Name
	}
	if request.Status != nil {
		team.Status = *request.Status
	}
	err = db.UpdateTeam(ctx, team)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK,
		utils.MsgSuccess,
		map[string]interface{}{
			"team": team,
		})
}

// ListTeams handles HTTP request, parses query params into db.TeamFilter and returns paginated response.
func ListTeams(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// authorization example
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	// build TeamFilter (reuses FilterOptions for sorting)
	filter, err := db.TeamFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}

	items, count, nextCursor, hasMore, err := db.ListTeams(ctx, filter)
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

func GetTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	team, err := db.GetTeamById(ctx, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if team == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTeamNotFound, nil)
	}
	teamSettings, err := db.GetTeamSettingsByTeamID(ctx, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if teamSettings == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTeamSettingsNotFound, nil)
	}
	return utils.SuccessResponse(http.StatusOK,
		utils.MsgSuccess,
		map[string]interface{}{
			"team":         team,
			"teamSettings": teamSettings,
		})
}

func DeleteTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	err := db.DeleteTeamByID(ctx, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccessTeamDeleted, nil)
}

func CreateTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var request CreateTeamRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	team, err := db.CreateTeam(ctx, request.Name)
	if err != nil {
		if err.Error() == "team already exists" {
			return utils.ErrorResponse(http.StatusBadRequest, utils.MsgErrorTeamExists, err)
		}
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	// Return the region from environment variable
	return utils.SuccessResponse(http.StatusOK,
		utils.MsgSuccessTeamCreated,
		map[string]interface{}{
			"team": team,
		})
}

func UploadTeamPicture(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	filename, ok := event.QueryStringParameters["filename"]
	if !ok || filename == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	contentType, ok := event.QueryStringParameters["contentType"]
	if !ok || contentType == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	presignedUrl, key, err := storage.GeneratePresignedUploadURLForTeamPicture(ctx, teamId, filename, contentType, 15)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	publicUrl := storage.GetPublicFileURL(key)
	err = db.UpdateTeamPicture(ctx, teamId, publicUrl)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK,
		utils.MsgSuccess,
		map[string]interface{}{
			"uploadUrl": presignedUrl,
			"key":       key,
			"fileUrl":   publicUrl,
		})
}
