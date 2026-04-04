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

type createRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type updateRoleRequest struct {
	Permissions []string `json:"permissions"`
}

func isTenantAuthorized(ctx context.Context, event events.APIGatewayProxyRequest, tenantId string) (bool, error) {
	if utils.IsAdmin(event.RequestContext.Authorizer) {
		return true, nil
	}
	return db.IsTenantAdmin(ctx, utils.GetCognitoUsername(event.RequestContext.Authorizer), tenantId)
}

func ListRoleDefinitions(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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
	tenantRoles, err := db.ListRoleDefinitionsByTenant(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	globalRoles, err := db.ListRoleDefinitionsByTenant(ctx, "global")
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	// Combine tenant and global roles. Global roles are represented in their
	// RoleDefinition (they typically have TenantId == "global" and IsDefault set).
	var respItems []*models.RoleDefinition
	respItems = append(respItems, tenantRoles...)
	respItems = append(respItems, globalRoles...)

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"items": respItems})
}

func CreateRoleDefinition(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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
	var req createRoleRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || req.Name == "" || len(req.Permissions) == 0 {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	role, err := db.CreateRoleDefinition(ctx, tenantId, req.Name, req.Permissions, false)
	if err != nil {
		log.WithError(err).Error("CreateRoleDefinition db error")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{"role": role})
}

func UpdateRoleDefinition(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	roleId := event.PathParameters["roleId"]
	if tenantId == "" || roleId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	existing, err := db.GetRoleDefinitionById(ctx, roleId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if existing == nil || existing.TenantId != tenantId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorRoleNotFound, nil)
	}
	if existing.IsDefault {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorRoleIsDefault, nil)
	}
	var req updateRoleRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || len(req.Permissions) == 0 {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	updated, err := db.UpdateRoleDefinitionPermissions(ctx, roleId, req.Permissions)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"role": updated})
}

func DeleteRoleDefinition(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	roleId := event.PathParameters["roleId"]
	if tenantId == "" || roleId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	existing, err := db.GetRoleDefinitionById(ctx, roleId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if existing == nil || existing.TenantId != tenantId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorRoleNotFound, nil)
	}
	if existing.IsDefault {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorRoleIsDefault, nil)
	}
	if err := db.DeleteRoleDefinition(ctx, roleId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}
