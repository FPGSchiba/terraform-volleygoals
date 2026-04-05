//go:build local

package utils

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
)

// This integration test seeds global role definitions (idempotent) into the
// local dev tables and asserts effective permission behaviour for member,
// trainer, admin and platform-admin scenarios. It is gated behind the `local`
// build tag so it only runs in a developer's local environment.
func TestSeededRolesAndEffectivePermissions(t *testing.T) {
	ctx := context.Background()
	db.InitClient(nil)

	// Seed global role definitions (idempotent)
	defs := models.GetDefinitions()

	adminPerms := []string{}
	for _, d := range defs {
		for _, a := range d.Actions {
			adminPerms = append(adminPerms, models.GetPermission(d.Id, a))
		}
	}

	trainerWrite := map[string]bool{"seasons": true, "goals": true, "progress_reports": true, "progress": true, "comments": true}
	trainerDelete := map[string]bool{"seasons": true, "goals": true, "progress_reports": true, "comments": true}
	trainerPerms := []string{}
	for _, d := range defs {
		trainerPerms = append(trainerPerms, models.GetPermission(d.Id, "read"))
		if trainerWrite[d.Id] {
			trainerPerms = append(trainerPerms, models.GetPermission(d.Id, "write"))
		}
		if trainerDelete[d.Id] {
			trainerPerms = append(trainerPerms, models.GetPermission(d.Id, "delete"))
		}
	}

	memberReadSet := map[string]bool{"teams": true, "members": true, "seasons": true, "goals": true, "progress_reports": true, "activities": true}
	memberPerms := []string{}
	for _, d := range defs {
		if memberReadSet[d.Id] {
			memberPerms = append(memberPerms, models.GetPermission(d.Id, "read"))
		}
	}

	roles := []struct {
		name  string
		perms []string
	}{
		{name: "admin", perms: adminPerms},
		{name: "trainer", perms: trainerPerms},
		{name: "member", perms: memberPerms},
		{name: "global_admin", perms: adminPerms},
	}

	for _, r := range roles {
		exists, err := db.GetRoleDefinitionByTenantAndName(ctx, "global", r.name)
		if err != nil {
			// If DynamoDB tables aren't present (dev environment not prepared),
			// skip the integration test rather than fail the whole suite.
			if strings.Contains(err.Error(), "Requested resource not found") || strings.Contains(err.Error(), "ResourceNotFoundException") {
				t.Skipf("skipping integration test because DynamoDB tables are not present: %v", err)
			}
			t.Fatalf("failed to check existing role %s: %v", r.name, err)
		}
		if exists == nil {
			if _, err := db.CreateRoleDefinition(ctx, "global", r.name, r.perms, true); err != nil {
				if strings.Contains(err.Error(), "Requested resource not found") || strings.Contains(err.Error(), "ResourceNotFoundException") {
					t.Skipf("skipping integration test because DynamoDB tables are not present: %v", err)
				}
				t.Fatalf("failed to create role %s: %v", r.name, err)
			}
		}
	}

	// Create a temporary team for the test
	teamName := "integ-test-team-" + time.Now().Format("20060102150405")
	team, err := db.CreateTeam(ctx, teamName)
	if err != nil {
		t.Fatalf("CreateTeam failed: %v", err)
	}
	defer func() {
		// best-effort cleanup
		_ = db.DeleteTeamByID(ctx, team.Id)
	}()

	// Add users with different roles
	memberUser := "integ-member-1"
	trainerUser := "integ-trainer-1"
	adminUser := "integ-admin-1"

	if _, err := db.AddTeamMember(ctx, team.Id, memberUser, models.TeamMemberRoleMember); err != nil {
		t.Fatalf("AddTeamMember member failed: %v", err)
	}
	if _, err := db.AddTeamMember(ctx, team.Id, trainerUser, models.TeamMemberRoleTrainer); err != nil {
		t.Fatalf("AddTeamMember trainer failed: %v", err)
	}
	if _, err := db.AddTeamMember(ctx, team.Id, adminUser, models.TeamMemberRoleAdmin); err != nil {
		t.Fatalf("AddTeamMember admin failed: %v", err)
	}

	// Quick permission assertions using HasTeamPermission helper which exercises
	// the full CheckPermission + global_admin fallback logic.
	// Member can read goals
	authorizerMember := map[string]interface{}{"claims": map[string]interface{}{"cognito:username": memberUser, "cognito:groups": []string{"user"}}}
	if !HasTeamPermission(ctx, authorizerMember, team.Id, models.Resource{Type: "goals"}, models.GetPermission("goals", "read")) {
		t.Fatalf("expected member to have read permission on goals")
	}
	// Member cannot write goals
	if HasTeamPermission(ctx, authorizerMember, team.Id, models.Resource{Type: "goals"}, models.GetPermission("goals", "write")) {
		t.Fatalf("expected member NOT to have write permission on goals")
	}

	// Trainer can write goals
	authorizerTrainer := map[string]interface{}{"claims": map[string]interface{}{"cognito:username": trainerUser, "cognito:groups": []string{"user"}}}
	if !HasTeamPermission(ctx, authorizerTrainer, team.Id, models.Resource{Type: "goals"}, models.GetPermission("goals", "write")) {
		t.Fatalf("expected trainer to have write permission on goals")
	}

	// Admin can delete goals
	authorizerAdmin := map[string]interface{}{"claims": map[string]interface{}{"cognito:username": adminUser, "cognito:groups": []string{"user"}}}
	if !HasTeamPermission(ctx, authorizerAdmin, team.Id, models.Resource{Type: "goals"}, models.GetPermission("goals", "delete")) {
		t.Fatalf("expected admin to have delete permission on goals")
	}

	// Platform admin (not a member) should be allowed via global_admin role
	platformAdminUser := "platform-admin-integ"
	authorizerPlatformAdmin := map[string]interface{}{"claims": map[string]interface{}{"cognito:username": platformAdminUser, "cognito:groups": []string{"admin"}}}
	// Should be allowed to perform a high-privilege action like delete
	if !HasTeamPermission(ctx, authorizerPlatformAdmin, team.Id, models.Resource{Type: "goals"}, models.GetPermission("goals", "delete")) {
		t.Fatalf("expected platform admin to have delete permission via global_admin role")
	}
}
