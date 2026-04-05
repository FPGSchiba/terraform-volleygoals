package resource_definitions

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

// GetResourceDefinitions returns a static list of resource definitions used by the frontend
func GetResourceDefinitions(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// Return the raw list (JSON array) expected by clients and tests.
	return utils.Response(http.StatusOK, models.GetDefinitions())
}
