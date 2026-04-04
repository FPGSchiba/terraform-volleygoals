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
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, GetDefinitions())
}

// GetDefinitions returns the static list of resource definitions for programmatic use.
func GetDefinitions() []models.ResourceDefinition {
	return []models.ResourceDefinition{
		{
			Id:                    "goals",
			Name:                  "Goals",
			Description:           "Individual or team goals",
			Actions:               []string{"read", "write", "delete"},
			AllowedChildResources: []string{"comments", "progressReports", "seasons"},
		},
		{
			Id:                    "comments",
			Name:                  "Comments",
			Description:           "Comments attached to goals or progress reports",
			Actions:               []string{"read", "write", "delete"},
			AllowedChildResources: []string{},
		},
		{
			Id:                    "progressReports",
			Name:                  "Progress Reports",
			Description:           "Reports of progress during a season",
			Actions:               []string{"read", "write", "delete"},
			AllowedChildResources: []string{"comments", "progress"},
		},
		{
			Id:                    "progress",
			Name:                  "Progress Entries",
			Description:           "Progress entries for reports",
			Actions:               []string{"read", "write", "delete"},
			AllowedChildResources: []string{},
		},
		{
			Id:                    "seasons",
			Name:                  "Seasons",
			Description:           "Team seasons",
			Actions:               []string{"read", "write", "delete"},
			AllowedChildResources: []string{"goals", "progressReports"},
		},
	}
}
