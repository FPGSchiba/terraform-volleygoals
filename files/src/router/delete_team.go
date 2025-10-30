//go:build deleteTeam

package router

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
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
		return nil, err
	}
	log.Printf("CreateTeamRequest: %v", request)
	// Return the region from environment variable
	return utils.Response(http.StatusOK,
		map[string]interface{}{
			"Region": os.Getenv("AWS_REGION"),
		})
}
