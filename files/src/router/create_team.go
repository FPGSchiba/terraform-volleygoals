//go:build createTeam

package router

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

type CreateTeamRequest struct {
	Name string `json:"name"`
}

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Printf("Event: %v", event)
	var request CreateTeamRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return nil, err // TODO: improve error handling
	}
	team, err := db.CreateTeam(context.Background(), request.Name)
	if err != nil {
		return nil, err // TODO: improve error handling
	}
	log.Printf("Created team: %v", team)
	// Return the region from environment variable
	return utils.Response(http.StatusOK,
		map[string]interface{}{
			"message": "Team created successfully",
			"team":    team,
		})
}
