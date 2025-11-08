package team_settings

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

func UpdateTeamSettings(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func GetTeamSettings(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	// TODO: Check if the user has access to the team
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
			"settings": teamSettings,
		})
}
