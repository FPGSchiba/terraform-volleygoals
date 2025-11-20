package team_settings

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

func UpdateTeamSettings(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	var request UpdateTeamSettingsRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	teamSettings, err := db.GetTeamSettingsByTeamID(ctx, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if teamSettings == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTeamSettingsNotFound, nil)
	}
	if request.AllowFileUploads != nil {
		teamSettings.AllowFileUploads = *request.AllowFileUploads
	}
	if request.AllowTeamGoalComments != nil {
		teamSettings.AllowTeamGoalComments = *request.AllowTeamGoalComments
	}
	if request.AllowIndividualGoalComments != nil {
		teamSettings.AllowIndividualGoalComments = *request.AllowIndividualGoalComments
	}
	if err := db.UpdateTeamSettings(ctx, teamSettings); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"teamSettings": teamSettings,
	})
}
