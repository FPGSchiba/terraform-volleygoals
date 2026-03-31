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
			log.Infof("role %q already exists, skipping", r.name)
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
		},
		{
			resourceType: models.ResourceTypeProgressReports,
			ownerPerms: []string{
				models.PermProgressReportsRead, models.PermProgressReportsWrite, models.PermProgressReportsDelete,
				models.PermCommentsRead, models.PermCommentsWrite,
			},
		},
		{
			resourceType: models.ResourceTypeProgress,
			ownerPerms:   []string{models.PermProgressRead, models.PermProgressWrite},
		},
		{
			resourceType:     models.ResourceTypeComments,
			ownerPerms:       []string{models.PermCommentsRead, models.PermCommentsWrite, models.PermCommentsDelete},
			parentOwnerPerms: []string{models.PermCommentsRead, models.PermCommentsWrite},
		},
	}

	for _, p := range policies {
		if _, err := db.UpsertOwnershipPolicy(ctx, "global", p.resourceType, p.ownerPerms, p.parentOwnerPerms); err != nil {
			return err
		}
		log.Infof("upserted ownership policy for %q", p.resourceType)
	}
	return nil
}
