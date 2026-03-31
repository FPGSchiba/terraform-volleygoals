package utils_test

import (
	"context"
	"testing"

	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers — build a PermissionChecker with injected loaders.

func memberLoader(role string) func(ctx context.Context, userID, teamID string) (*models.TeamMember, error) {
	return func(ctx context.Context, userID, teamID string) (*models.TeamMember, error) {
		return &models.TeamMember{Role: models.TeamMemberRole(role)}, nil
	}
}

func teamLoader(tenantId *string) func(ctx context.Context, teamID string) (*models.Team, error) {
	return func(ctx context.Context, teamID string) (*models.Team, error) {
		return &models.Team{Id: teamID, TenantId: tenantId}, nil
	}
}

func ownershipLoader(ownerPerms, parentPerms []string) func(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error) {
	return func(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error) {
		return &models.OwnershipPolicy{
			OwnerPermissions:       ownerPerms,
			ParentOwnerPermissions: parentPerms,
		}, nil
	}
}

func roleLoader(perms []string) func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
	return func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
		return &models.RoleDefinition{Permissions: perms}, nil
	}
}

func nilRoleLoader() func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
	return func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
		return nil, nil
	}
}

func TestCheckPermission_OwnerCanReadOwnGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("member"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete}, nil),
		LoadRoleByTenant: nilRoleLoader(),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-1"}
	allowed, err := checker.Check(context.Background(), "user-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.True(t, allowed, "owner should be able to read their own goal")
}

func TestCheckPermission_NonOwnerMemberCannotReadGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("member"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead}, nil),
		LoadRoleByTenant: nilRoleLoader(),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-2"}
	allowed, err := checker.Check(context.Background(), "user-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.False(t, allowed, "non-owner member should not read another member's goal")
}

func TestCheckPermission_TrainerCanReadAnyGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("trainer"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader(nil, nil),
		LoadRoleByTenant: roleLoader([]string{models.PermGoalsRead, models.PermGoalsWrite}),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-2"}
	allowed, err := checker.Check(context.Background(), "trainer-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.True(t, allowed, "trainer role should be able to read any goal")
}

func TestCheckPermission_TrainerCannotEditTeam(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("trainer"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader(nil, nil),
		LoadRoleByTenant: roleLoader([]string{models.PermTeamsRead}), // trainer has read, not write
	}
	resource := models.Resource{Type: models.ResourceTypeTeams}
	allowed, err := checker.Check(context.Background(), "trainer-1", "team-1", resource, models.PermTeamsWrite)
	require.NoError(t, err)
	assert.False(t, allowed, "trainer should not be able to edit the team")
}

func TestCheckPermission_AdminCannotReadOtherMemberGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("admin"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead}, nil),
		LoadRoleByTenant: roleLoader([]string{models.PermTeamsRead, models.PermTeamsWrite, models.PermTeamsDelete}), // admin has no goals:read
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-other"}
	allowed, err := checker.Check(context.Background(), "admin-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.False(t, allowed, "admin should not read another member's goal")
}

func TestCheckPermission_AdminCanReadOwnGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("admin"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete}, nil),
		LoadRoleByTenant: roleLoader([]string{models.PermTeamsWrite}),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "admin-1"}
	allowed, err := checker.Check(context.Background(), "admin-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.True(t, allowed, "admin should read their own goal via ownership")
}

func TestCheckPermission_GoalOwnerCanReadCommentViaParentOwnership(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("member"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermCommentsRead}, []string{models.PermCommentsRead, models.PermCommentsWrite}),
		LoadRoleByTenant: nilRoleLoader(),
	}
	// comment.OwnedBy = trainer, comment.ParentOwnedBy = member (goal owner)
	resource := models.Resource{
		Type:          models.ResourceTypeComments,
		OwnedBy:       "trainer-1",
		ParentOwnedBy: "member-1",
	}
	allowed, err := checker.Check(context.Background(), "member-1", "team-1", resource, models.PermCommentsRead)
	require.NoError(t, err)
	assert.True(t, allowed, "goal owner should read comments on their goal via parent ownership")
}

func TestCheckPermission_DenyByDefault(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("member"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader(nil, nil),
		LoadRoleByTenant: nilRoleLoader(),
	}
	resource := models.Resource{Type: models.ResourceTypeTeams}
	allowed, err := checker.Check(context.Background(), "member-1", "team-1", resource, models.PermTeamsDelete)
	require.NoError(t, err)
	assert.False(t, allowed, "should deny by default when no policy matches")
}

func TestCheckPermission_NilTeamMemberDenies(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember: func(ctx context.Context, userID, teamID string) (*models.TeamMember, error) {
			return nil, nil // not a team member
		},
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead}, nil),
		LoadRoleByTenant: nilRoleLoader(),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-1"}
	allowed, err := checker.Check(context.Background(), "user-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.False(t, allowed, "non-member should be denied even if they own the resource")
}
