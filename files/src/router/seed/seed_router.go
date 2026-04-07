package seed

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

// SeedDefaults seeds global RoleDefinition and OwnershipPolicy records.
// Invoked by Terraform via aws_lambda_invocation after tables are created.
// Safe to re-run: existing records are skipped.
func SeedDefaults(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if err := seedRoleDefinitions(ctx); err != nil {
		log.WithError(err).Error("seed role definitions failed")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if err := seedOwnershipPolicies(ctx); err != nil {
		log.WithError(err).Error("seed ownership policies failed")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]string{"status": "seeded"})
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
	trainerWrite := map[string]bool{"seasons": true, "team_goals": true, "progress_reports": true, "progress": true, "comments": true}
	trainerDelete := map[string]bool{"seasons": true, "team_goals": true, "progress_reports": true, "comments": true}
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
	memberReadSet := map[string]bool{"teams": true, "members": true, "seasons": true, "team_goals": true, "progress_reports": true, "activities": true}
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
		// platform-level role stored under the "global" tenant
		{name: "global_admin", permissions: adminPerms},
		{name: "tenant_admin", permissions: adminPerms},
	}

	for _, r := range roles {
		existing, err := db.GetRoleDefinitionByTenantAndName(ctx, "global", r.name)
		if err != nil {
			return err
		}
		if existing != nil {
			// Update the existing role with new permissions if they differ
			if _, err := db.UpdateRoleDefinitionPermissions(ctx, existing.Id, r.permissions); err != nil {
				return err
			}
			log.Infof("updated role %q", r.name)
			continue
		}
		if _, err := db.CreateRoleDefinition(ctx, "global", r.name, r.permissions, true); err != nil {
			return err
		}
		log.Infof("created role %q", r.name)
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

		_, err := db.UpsertOwnershipPolicy(ctx, "global", d.Id, ownerPerms, parentOwnerPerms)
		if err != nil {
			return err
		}
		log.Infof("upserted ownership policy for %q", d.Id)
	}
	return nil
}
