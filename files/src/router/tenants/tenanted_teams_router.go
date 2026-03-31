package tenants

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

type createTenantedTeamRequest struct {
	Name string `json:"name"`
}

// CreateTenantedTeam creates a new team already linked to the given tenant.
// Authorization: global ADMINS or tenant admin.
func CreateTenantedTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	tenant, err := db.GetTenantById(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if tenant == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantNotFound, nil)
	}
	var req createTenantedTeamRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || req.Name == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	team, err := db.CreateTeamWithTenant(ctx, req.Name, tenantId)
	if err != nil {
		log.WithError(err).Error("CreateTenantedTeam db error")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{"team": team})
}
