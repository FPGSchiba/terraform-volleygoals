//go:build createTeam

package router

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

type CreateTeamRequest struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var request CreateTeamRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	team, err := db.CreateTeam(context.Background(), request.Name)
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
