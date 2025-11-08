package teams

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

func UpdateTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func ListTeams(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func GetTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	team, err := db.GetTeamByID(ctx, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if team == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTeamNotFound, nil)
	}
	return utils.SuccessResponse(http.StatusOK,
		utils.MsgSuccess,
		map[string]interface{}{
			"team": team,
		})
}

func DeleteTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
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
