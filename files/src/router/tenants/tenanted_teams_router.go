package tenants

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
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

func ListTenantedTeams(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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

	// Build TeamFilter from query params
	filter, err := db.TeamFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}

	items, count, nextCursor, hasMore, err := db.ListTeamsByTenant(ctx, tenantId, filter)
	if err != nil {
		log.WithError(err).Error("ListTenantedTeams db error")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	nextToken := ""
	if nextCursor != nil {
		nextToken, err = models.EncodeCursor(nextCursor)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
	}

	resp := models.PaginationResponse{
		Items:     items,
		Count:     count,
		NextToken: nextToken,
		HasMore:   hasMore,
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items":     resp.Items,
		"count":     resp.Count,
		"nextToken": resp.NextToken,
		"hasMore":   resp.HasMore,
	})
}
