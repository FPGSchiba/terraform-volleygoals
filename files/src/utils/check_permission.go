package utils

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
)

// PermissionChecker evaluates whether an actor is allowed to perform an action
// on a resource within a team. Loader functions are injected for testability.
type PermissionChecker struct {
	LoadTeamMember   func(ctx context.Context, userID, teamID string) (*models.TeamMember, error)
	LoadTeam         func(ctx context.Context, teamID string) (*models.Team, error)
	LoadOwnership    func(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error)
	LoadRoleByTenant func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error)
}

// DefaultChecker wires the checker to the real db package.
var DefaultChecker = &PermissionChecker{
	LoadTeamMember:   db.GetTeamMemberByUserIDAndTeamID,
	LoadTeam:         db.GetTeamById,
	LoadOwnership:    db.GetOwnershipPolicy,
	LoadRoleByTenant: db.GetRoleDefinitionByTenantAndName,
}

// CheckPermission is the single entry point for all team-level permission
// checks. It replaces IsTeamAdmin, IsTeamTrainer, HasTeamAccess, etc.
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

	// Step 4: tenant-specific role definition
	if tenantId != "" {
		roleDef, err := pc.LoadRoleByTenant(ctx, tenantId, string(member.Role))
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
