//go:build ignore

// seed_defaults seeds global RoleDefinition and OwnershipPolicy records into
// DynamoDB. Run once per environment:
//
//	go run -tags local files/src/scripts/seed_defaults/main.go
//
// The -tags local flag uses db/local_init.go with hardcoded dev-* table names.
package main

import (
	"context"
	"log"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
)

func main() {
	ctx := context.Background()

	db.InitClient(nil)

	if err := seedRoleDefinitions(ctx); err != nil {
		log.Fatalf("seed role definitions: %v", err)
	}
	if err := seedOwnershipPolicies(ctx); err != nil {
		log.Fatalf("seed ownership policies: %v", err)
	}
	log.Println("Seed complete")
}

func seedRoleDefinitions(ctx context.Context) error {
	roles := []struct {
		name        string
		permissions []string
	}{
		{
			name: "admin",
			permissions: []string{
				models.PermTeamsRead, models.PermTeamsWrite, models.PermTeamsDelete,
				models.PermTeamSettingsRead, models.PermTeamSettingsWrite,
				models.PermMembersRead, models.PermMembersWrite, models.PermMembersDelete,
				models.PermInvitesRead, models.PermInvitesWrite, models.PermInvitesDelete,
				models.PermSeasonsRead,
				models.PermActivitiesRead,
			},
		},
		{
			name: "trainer",
			permissions: []string{
				models.PermTeamsRead,
				models.PermTeamSettingsRead,
				models.PermMembersRead,
				models.PermSeasonsRead, models.PermSeasonsWrite, models.PermSeasonsDelete,
				models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete,
				models.PermProgressReportsRead, models.PermProgressReportsWrite, models.PermProgressReportsDelete,
				models.PermProgressRead, models.PermProgressWrite,
				models.PermCommentsRead, models.PermCommentsWrite, models.PermCommentsDelete,
				models.PermActivitiesRead,
			},
		},
		{
			name: "member",
			permissions: []string{
				models.PermTeamsRead,
				models.PermMembersRead,
				models.PermSeasonsRead,
			},
		},
	}

	for _, r := range roles {
		existing, err := db.GetRoleDefinitionByTenantAndName(ctx, "global", r.name)
		if err != nil {
			return err
		}
		if existing != nil {
			log.Printf("  role %q already exists, skipping", r.name)
			continue
		}
		def, err := db.CreateRoleDefinition(ctx, "global", r.name, r.permissions, true)
		if err != nil {
			return err
		}
		log.Printf("  created role %q (%s)", def.Name, def.Id)
	}
	return nil
}

func seedOwnershipPolicies(ctx context.Context) error {
	policies := []struct {
		resourceType     string
		ownerPerms       []string
		parentOwnerPerms []string
	}{
		{
			resourceType: models.ResourceTypeGoals,
			ownerPerms: []string{
				models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete,
				models.PermCommentsRead, models.PermCommentsWrite,
			},
			parentOwnerPerms: nil,
		},
		{
			resourceType: models.ResourceTypeProgressReports,
			ownerPerms: []string{
				models.PermProgressReportsRead, models.PermProgressReportsWrite, models.PermProgressReportsDelete,
				models.PermCommentsRead, models.PermCommentsWrite,
			},
			parentOwnerPerms: nil,
		},
		{
			resourceType: models.ResourceTypeProgress,
			ownerPerms:   []string{models.PermProgressRead, models.PermProgressWrite},
			parentOwnerPerms: nil,
		},
		{
			resourceType: models.ResourceTypeComments,
			ownerPerms:   []string{models.PermCommentsRead, models.PermCommentsWrite, models.PermCommentsDelete},
			parentOwnerPerms: []string{models.PermCommentsRead, models.PermCommentsWrite},
		},
	}

	for _, p := range policies {
		policy, err := db.UpsertOwnershipPolicy(ctx, "global", p.resourceType, p.ownerPerms, p.parentOwnerPerms)
		if err != nil {
			return err
		}
		log.Printf("  upserted ownership policy for %q (%s)", policy.ResourceType, policy.Id)
	}
	return nil
}
