//go:build getTeams

package router

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/utils"
)

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Return the region from environment variable
	return utils.Response(http.StatusOK,
		map[string]interface{}{
			"Region": os.Getenv("AWS_REGION"),
		})
}
