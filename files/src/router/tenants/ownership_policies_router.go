package tenants

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	rd "github.com/fpgschiba/volleygoals/router/resource_definitions"
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
	// Return a policy per resource type. If a tenant-specific policy is missing
	// for a resource type, fall back to the global policy (handled by GetOwnershipPolicy).
	resourceTypes := []string{"goals", "comments", "progressReports", "progress", "seasons"}

	type policyOut struct {
		Id                     string   `json:"id,omitempty"`
		TenantId               string   `json:"tenantId,omitempty"`
		ResourceType           string   `json:"resourceType"`
		OwnerPermissions       []string `json:"ownerPermissions"`
		ParentOwnerPermissions []string `json:"parentOwnerPermissions"`
	}

	// Build flat list
	var flat []*policyOut
	// Build nested map if requested
	nested := make(map[string]*policyOut)

	for _, rt := range resourceTypes {
		p, err := db.GetOwnershipPolicy(ctx, tenantId, rt)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		po := &policyOut{ResourceType: rt}
		if p != nil {
			po.Id = p.Id
			po.TenantId = p.TenantId
			po.OwnerPermissions = p.OwnerPermissions
			po.ParentOwnerPermissions = p.ParentOwnerPermissions
		}
		flat = append(flat, po)
		nested[rt] = po
	}

	// If format=nested requested, return nested map for backward compatibility
	if event.QueryStringParameters != nil {
		if v, ok := event.QueryStringParameters["format"]; ok && v == "nested" {
			return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"policies": nested})
		}
	}

	// Default: return normalized flat list
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"message": utils.MsgSuccess, "policies": flat})
}

func UpdateOwnershipPolicy(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	resourceType := event.PathParameters["resourceType"]
	if tenantId == "" || resourceType == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	// If caller attempts to edit the global policy, we create a tenant-specific
	// override instead. The caller must provide the target tenant via the
	// "X-Target-Tenant" header or the "targetTenant" query parameter.
	effectiveTenant := tenantId
	if tenantId == "global" {
		// Prefer explicit header, fall back to query param
		if v, ok := event.Headers["X-Target-Tenant"]; ok && v != "" {
			effectiveTenant = v
		} else if qp, ok := event.QueryStringParameters["targetTenant"]; ok && qp != "" {
			effectiveTenant = qp
		} else {
			return utils.ErrorResponse(http.StatusBadRequest, "when editing global policy, targetTenant must be provided via X-Target-Tenant header or targetTenant query param", nil)
		}
	}

	ok, err := isTenantAuthorized(ctx, event, effectiveTenant)
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

	// Validate permission format and actions
	defs := rd.GetDefinitions()
	// build map resource -> set(actions)
	actionMap := make(map[string]map[string]bool)
	for _, d := range defs {
		m := make(map[string]bool)
		for _, a := range d.Actions {
			m[a] = true
		}
		actionMap[d.Id] = m
	}

	validatePerms := func(perms []string) error {
		for _, p := range perms {
			if strings.TrimSpace(p) == "" {
				continue
			}
			parts := strings.SplitN(p, ":", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid permission format: %s", p)
			}
			res := parts[0]
			act := parts[1]
			if acts, ok := actionMap[res]; ok {
				if !acts[act] {
					return fmt.Errorf("invalid action '%s' for resource '%s'", act, res)
				}
			} else {
				return fmt.Errorf("unknown resource in permission: %s", res)
			}
		}
		return nil
	}

	if err := validatePerms(req.OwnerPermissions); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	if err := validatePerms(req.ParentOwnerPermissions); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	policy, err := db.UpsertOwnershipPolicy(ctx, effectiveTenant, resourceType, req.OwnerPermissions, req.ParentOwnerPermissions)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"policy": policy})
}

// BatchUpsertOwnershipPolicies accepts an array of policies and applies them.
// It attempts a compensation strategy on error: deletes newly created policies
// and restores previous UpdatedAt/UpdatedTo where possible.
func BatchUpsertOwnershipPolicies(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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

	// Parse incoming array
	var items []struct {
		ResourceType           string   `json:"resourceType"`
		OwnerPermissions       []string `json:"ownerPermissions"`
		ParentOwnerPermissions []string `json:"parentOwnerPermissions"`
	}
	if err := json.Unmarshal([]byte(event.Body), &items); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}

	// Prepare validation map
	defs := rd.GetDefinitions()
	actionMap := make(map[string]map[string]bool)
	for _, d := range defs {
		m := make(map[string]bool)
		for _, a := range d.Actions {
			m[a] = true
		}
		actionMap[d.Id] = m
	}
	validatePerms := func(perms []string) error {
		for _, p := range perms {
			if strings.TrimSpace(p) == "" {
				continue
			}
			parts := strings.SplitN(p, ":", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid permission format: %s", p)
			}
			res := parts[0]
			act := parts[1]
			if acts, ok := actionMap[res]; ok {
				if !acts[act] {
					return fmt.Errorf("invalid action '%s' for resource '%s'", act, res)
				}
			} else {
				return fmt.Errorf("unknown resource in permission: %s", res)
			}
		}
		return nil
	}

	// Track created policies and previous states for compensation
	var createdIDs []string
	prevByResource := make(map[string]*models.OwnershipPolicy)
	var createdPolicies []*models.OwnershipPolicy

	for _, it := range items {
		if it.ResourceType == "" {
			return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, fmt.Errorf("resourceType required"))
		}
		if err := validatePerms(it.OwnerPermissions); err != nil {
			return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
		}
		if err := validatePerms(it.ParentOwnerPermissions); err != nil {
			return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
		}

		newPolicy, prev, err := db.UpsertOwnershipPolicyReturnPrev(ctx, tenantId, it.ResourceType, it.OwnerPermissions, it.ParentOwnerPermissions)
		if err != nil {
			// Compensation: best-effort
			_ = db.CompensateUpserts(ctx, createdIDs, prevByResource)
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		createdPolicies = append(createdPolicies, newPolicy)
		createdIDs = append(createdIDs, newPolicy.Id)
		if prev != nil {
			prevByResource[it.ResourceType] = prev
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"policies": createdPolicies})
}

// GetResourceModel returns resource definitions and effective policies in one call.
func GetResourceModel(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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

	defs := rd.GetDefinitions()
	resourceTypes := []string{"goals", "comments", "progressReports", "progress", "seasons"}

	type policyOut struct {
		Id                     string   `json:"id,omitempty"`
		TenantId               string   `json:"tenantId,omitempty"`
		ResourceType           string   `json:"resourceType"`
		OwnerPermissions       []string `json:"ownerPermissions"`
		ParentOwnerPermissions []string `json:"parentOwnerPermissions"`
	}

	var policies []*policyOut
	for _, rt := range resourceTypes {
		p, err := db.GetOwnershipPolicy(ctx, tenantId, rt)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		po := &policyOut{ResourceType: rt}
		if p != nil {
			po.Id = p.Id
			po.TenantId = p.TenantId
			po.OwnerPermissions = p.OwnerPermissions
			po.ParentOwnerPermissions = p.ParentOwnerPermissions
		}
		policies = append(policies, po)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"resourceDefinitions": defs, "policies": policies})
}

// PreviewEffectivePermissions returns effective permissions for a resource type/id
func PreviewEffectivePermissions(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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

	var body struct {
		ResourceType string `json:"resourceType"`
		ResourceId   string `json:"resourceId,omitempty"`
	}
	if err := json.Unmarshal([]byte(event.Body), &body); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	if body.ResourceType == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, fmt.Errorf("resourceType required"))
	}

	// Validate resourceType exists
	defs := rd.GetDefinitions()
	found := false
	for _, d := range defs {
		if d.Id == body.ResourceType {
			found = true
			break
		}
	}
	if !found {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, fmt.Errorf("unknown resourceType: %s", body.ResourceType))
	}

	p, err := db.GetOwnershipPolicy(ctx, tenantId, body.ResourceType)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	var effective []string
	var sources []map[string]string
	if p != nil {
		seen := make(map[string]bool)
		for _, perm := range p.OwnerPermissions {
			if !seen[perm] {
				effective = append(effective, perm)
				sources = append(sources, map[string]string{"permission": perm, "sourceResourceType": p.ResourceType, "policyId": p.Id})
				seen[perm] = true
			}
		}
		for _, perm := range p.ParentOwnerPermissions {
			if !seen[perm] {
				effective = append(effective, perm)
				sources = append(sources, map[string]string{"permission": perm, "sourceResourceType": p.ResourceType, "policyId": p.Id})
				seen[perm] = true
			}
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"effectivePermissions": effective, "sources": sources})
}
