package utils

import (
	"context"
	"fmt"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
)

// PermissionChecker evaluates whether an actor is allowed to perform an action
// on a resource within a team. Loader functions are injected for testability.
type PermissionChecker struct {
	LoadTeamMember        func(ctx context.Context, userID, teamID string) (*models.TeamMember, error)
	LoadTeam              func(ctx context.Context, teamID string) (*models.Team, error)
	LoadOwnership         func(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error)
	LoadRoleByTenantExact func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error)
	LoadRoleByTenant      func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error)
}

// DefaultChecker wires the checker to the real db package.
var DefaultChecker = &PermissionChecker{
	LoadTeamMember:        db.GetTeamMemberByUserIDAndTeamID,
	LoadTeam:              db.GetTeamById,
	LoadOwnership:         db.GetOwnershipPolicy,
	LoadRoleByTenantExact: db.GetRoleDefinitionByTenantExact,
	LoadRoleByTenant:      db.GetRoleDefinitionByTenantAndName,
}

// PreloadedData holds pre-fetched permission data for a single request.
type PreloadedData struct {
	Member          *models.TeamMember
	Team            *models.Team
	OwnershipByType map[string]*models.OwnershipPolicy
	RoleExact       *models.RoleDefinition
	RoleGlobal      *models.RoleDefinition
}

// PreloadPermissions fetches all permission data needed to check activities.
func PreloadPermissions(ctx context.Context, actorId, teamId string) (*PreloadedData, error) {
	member, err := db.GetTeamMemberByUserIDAndTeamID(ctx, actorId, teamId)
	if err != nil {
		return nil, fmt.Errorf("PreloadPermissions: load member: %w", err)
	}
	if member == nil {
		return &PreloadedData{}, nil
	}

	team, err := db.GetTeamById(ctx, teamId)
	if err != nil {
		return nil, fmt.Errorf("PreloadPermissions: load team: %w", err)
	}

	tenantId := ""
	if team != nil && team.TenantId != nil {
		tenantId = *team.TenantId
	}

	resourceTypes := []string{
		models.ResourceTypeGoals,
		models.ResourceTypeComments,
		models.ResourceTypeProgressReports,
		models.ResourceTypeProgress,
		models.ResourceTypeSeasons,
	}
	ownershipByType := make(map[string]*models.OwnershipPolicy, len(resourceTypes))
	for _, rt := range resourceTypes {
		policy, perr := db.GetOwnershipPolicy(ctx, tenantId, rt)
		if perr != nil {
			return nil, fmt.Errorf("PreloadPermissions: load ownership policy for %s: %w", rt, perr)
		}
		ownershipByType[rt] = policy
	}

	var roleExact, roleGlobal *models.RoleDefinition
	roleName := string(member.Role)
	if tenantId != "" {
		roleExact, err = db.GetRoleDefinitionByTenantExact(ctx, tenantId, roleName)
		if err != nil {
			return nil, fmt.Errorf("PreloadPermissions: load tenant role: %w", err)
		}
	}
	roleGlobal, err = db.GetRoleDefinitionByTenantAndName(ctx, "global", roleName)
	if err != nil {
		return nil, fmt.Errorf("PreloadPermissions: load global role: %w", err)
	}

	return &PreloadedData{
		Member:          member,
		Team:            team,
		OwnershipByType: ownershipByType,
		RoleExact:       roleExact,
		RoleGlobal:      roleGlobal,
	}, nil
}

// CanReadActivity returns true if the actor can read the given activity.
func (pd *PreloadedData) CanReadActivity(actorId string, a *models.Activity) bool {
	if pd.Member == nil {
		return false
	}

	readPerm := resourceReadPerm(a.TargetType)
	resource := models.Resource{
		Type:    a.TargetType,
		OwnedBy: a.TargetOwnerId,
	}

	if resource.OwnedBy != "" && resource.OwnedBy == actorId {
		if policy, ok := pd.OwnershipByType[resource.Type]; ok && policy != nil {
			if containsString(policy.OwnerPermissions, readPerm) {
				return true
			}
		}
	}

	if pd.RoleExact != nil && containsString(pd.RoleExact.Permissions, readPerm) {
		return true
	}

	if pd.RoleGlobal != nil && containsString(pd.RoleGlobal.Permissions, readPerm) {
		return true
	}

	return false
}

func resourceReadPerm(targetType string) string {
	switch targetType {
	case models.ResourceTypeGoals:
		return models.PermGoalsRead
	case models.ResourceTypeComments:
		return models.PermCommentsRead
	case models.ResourceTypeProgressReports:
		return models.PermProgressReportsRead
	case models.ResourceTypeProgress:
		return models.PermProgressRead
	case models.ResourceTypeSeasons:
		return models.PermSeasonsRead
	default:
		return ""
	}
}

// CheckPermission is the single entry point for all team-level permission checks.
func CheckPermission(ctx context.Context, actorId, teamId string, resource models.Resource, action string) (bool, error) {
	return DefaultChecker.Check(ctx, actorId, teamId, resource, action)
}

// Check runs the evaluation chain against the injected loaders.
func (pc *PermissionChecker) Check(ctx context.Context, actorId, teamId string, resource models.Resource, action string) (bool, error) {
	// Step 1: actor must be an active team member
	member, err := pc.LoadTeamMember(ctx, actorId, teamId)
	if err != nil {
		return false, err
	}
	if member == nil {
		return false, nil
	}

	// Resolve tenantId from team (empty string = no tenant = use global defaults only)
	team, err := pc.LoadTeam(ctx, teamId)
	if err != nil {
		return false, err
	}
	tenantId := ""
	if team != nil && team.TenantId != nil {
		tenantId = *team.TenantId
	}

	// Step 2: direct ownership check
	if resource.OwnedBy != "" && resource.OwnedBy == actorId {
		policy, err := pc.LoadOwnership(ctx, tenantId, resource.Type)
		if err != nil {
			return false, err
		}
		if policy != nil && containsString(policy.OwnerPermissions, action) {
			return true, nil
		}
	}

	// Step 3: parent ownership check
	if resource.ParentOwnedBy != "" && resource.ParentOwnedBy == actorId {
		policy, err := pc.LoadOwnership(ctx, tenantId, resource.Type)
		if err != nil {
			return false, err
		}
		if policy != nil && containsString(policy.ParentOwnerPermissions, action) {
			return true, nil
		}
	}

	// Step 4: tenant-specific role definition (exact match only, no global fallback)
	if tenantId != "" {
		roleDef, err := pc.LoadRoleByTenantExact(ctx, tenantId, string(member.Role))
		if err != nil {
			return false, err
		}
		if roleDef != nil && containsString(roleDef.Permissions, action) {
			return true, nil
		}
	}

	// Step 5: global default role definition
	roleDef, err := pc.LoadRoleByTenant(ctx, "global", string(member.Role))
	if err != nil {
		return false, err
	}
	if roleDef != nil && containsString(roleDef.Permissions, action) {
		return true, nil
	}

	return false, nil
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
