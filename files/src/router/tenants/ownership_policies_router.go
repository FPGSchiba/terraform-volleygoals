package tenants

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

type updateOwnershipPolicyRequest struct {
	OwnerPermissions       []string `json:"ownerPermissions"`
	ParentOwnerPermissions []string `json:"parentOwnerPermissions"`
}

func ListOwnershipPolicies(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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
	policies, err := db.ListOwnershipPoliciesByTenant(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"policies": policies})
}

func UpdateOwnershipPolicy(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	resourceType := event.PathParameters["resourceType"]
	if tenantId == "" || resourceType == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var req updateOwnershipPolicyRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	policy, err := db.UpsertOwnershipPolicy(ctx, tenantId, resourceType, req.OwnerPermissions, req.ParentOwnerPermissions)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"policy": policy})
}
