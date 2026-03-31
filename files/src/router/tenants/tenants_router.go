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

type createTenantRequest struct {
	Name string `json:"name"`
}

type updateTenantRequest struct {
	Name *string `json:"name,omitempty"`
}

type addTenantMemberRequest struct {
	UserId string                  `json:"userId"`
	Role   models.TenantMemberRole `json:"role"`
}

func CreateTenant(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var req createTenantRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || req.Name == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	ownerId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	tenant, err := db.CreateTenant(ctx, req.Name, ownerId)
	if err != nil {
		log.WithError(err).Error("CreateTenant db error")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{"tenant": tenant})
}

func GetTenant(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
		member, err := db.GetTenantMemberByUserAndTenant(ctx, userId, tenantId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if member == nil || member.Status != models.TenantMemberStatusActive {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	tenant, err := db.GetTenantById(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if tenant == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantNotFound, nil)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"tenant": tenant})
}

func UpdateTenant(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		isTA, err := db.IsTenantAdmin(ctx, utils.GetCognitoUsername(event.RequestContext.Authorizer), tenantId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if !isTA {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	tenant, err := db.GetTenantById(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if tenant == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantNotFound, nil)
	}
	var req updateTenantRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if err := db.UpdateTenant(ctx, tenant); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"tenant": tenant})
}

func DeleteTenant(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	tenant, err := db.GetTenantById(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if tenant == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantNotFound, nil)
	}
	if err := db.DeleteTenantById(ctx, tenantId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}

func AddTenantMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		isTA, err := db.IsTenantAdmin(ctx, utils.GetCognitoUsername(event.RequestContext.Authorizer), tenantId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if !isTA {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	var req addTenantMemberRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || req.UserId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	if req.Role == "" {
		req.Role = models.TenantMemberRoleMember
	}
	member, err := db.AddTenantMember(ctx, tenantId, req.UserId, req.Role)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{"member": member})
}

func RemoveTenantMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	memberId := event.PathParameters["memberId"]
	if tenantId == "" || memberId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		isTA, err := db.IsTenantAdmin(ctx, utils.GetCognitoUsername(event.RequestContext.Authorizer), tenantId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if !isTA {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	member, err := db.GetTenantMemberById(ctx, memberId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if member == nil || member.TenantId != tenantId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantMemberNotFound, nil)
	}
	if err := db.RemoveTenantMember(ctx, memberId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}
