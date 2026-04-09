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
	defs := models.GetDefinitions()

	// Admin: full access to all resource actions
	adminPerms := []string{}
	for _, d := range defs {
		for _, a := range d.Actions {
			adminPerms = append(adminPerms, models.GetPermission(d.Id, a))
		}
	}

	// Trainer: read across all resources, write/delete for content-related resources
	trainerWrite := map[string]bool{"seasons": true, "goals": true, "progress_reports": true, "progress": true, "comments": true}
	trainerDelete := map[string]bool{"seasons": true, "goals": true, "progress_reports": true, "comments": true}
	trainerPerms := []string{}
	for _, d := range defs {
		// read for all
		trainerPerms = append(trainerPerms, models.GetPermission(d.Id, "read"))
		if trainerWrite[d.Id] {
			trainerPerms = append(trainerPerms, models.GetPermission(d.Id, "write"))
		}
		if trainerDelete[d.Id] {
			trainerPerms = append(trainerPerms, models.GetPermission(d.Id, "delete"))
		}
	}

	// Member: conservative read access to team-scoped resources
	memberReadSet := map[string]bool{"teams": true, "members": true, "seasons": true, "goals": true, "progress_reports": true, "activities": true}
	memberPerms := []string{}
	for _, d := range defs {
		if memberReadSet[d.Id] {
			memberPerms = append(memberPerms, models.GetPermission(d.Id, "read"))
		}
	}

	roles := []struct {
		name        string
		permissions []string
	}{
		// tenant-scoped roles
		{name: "admin", permissions: adminPerms},
		{name: "trainer", permissions: trainerPerms},
		{name: "member", permissions: memberPerms},
		// platform-level role stored under the "global" tenant. This role is
		// intended to represent platform administrators and should include
		// every available permission. Seeder is idempotent so if the role
		// already exists it will be skipped.
		{name: "global_admin", permissions: adminPerms},
		{name: "tenant_admin", permissions: adminPerms},
	}

	for _, r := range roles {
		existing, err := db.GetRoleDefinitionByTenantAndName(ctx, "global", r.name)
		if err != nil {
			return err
		}
		if existing != nil {
			if _, err := db.UpdateRoleDefinitionPermissions(ctx, existing.Id, r.permissions); err != nil {
				return err
			}
			log.Printf("  updated role %q (%s)", existing.Name, existing.Id)
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
	defs := models.GetDefinitions()

	for _, d := range defs {
		// Owner permissions: all actions on the resource
		ownerPerms := []string{}
		for _, a := range d.Actions {
			ownerPerms = append(ownerPerms, models.GetPermission(d.Id, a))
		}
		// If resource supports activities as children, owners should be able to read activities
		for _, c := range d.AllowedChildResources {
			if c == "activities" {
				ownerPerms = append(ownerPerms, models.GetPermission("activities", "read"))
				break
			}
		}

		var parentOwnerPerms []string
		// Comments: allow parent owner (e.g., goal/report owner) to read/write comments
		if d.Id == "comments" {
			parentOwnerPerms = []string{models.GetPermission("comments", "read"), models.GetPermission("comments", "write")}
		}

		policy, err := db.UpsertOwnershipPolicy(ctx, "global", d.Id, ownerPerms, parentOwnerPerms)
		if err != nil {
			return err
		}
		log.Printf("  upserted ownership policy for %q (%s)", policy.ResourceType, policy.Id)
	}
	return nil
}
