package models

// TenantIDGlobal is the sentinel tenantId value for global default policies.
// DynamoDB cannot index null values, so "global" is used instead of an empty string.
const TenantIDGlobal = "global"

// Permission constants — resource:action pairs used in RoleDefinition.Permissions
// and OwnershipPolicy.OwnerPermissions.
const (
	PermTeamsRead   = "teams:read"
	PermTeamsWrite  = "teams:write"
	PermTeamsDelete = "teams:delete"

	PermTeamSettingsRead  = "team_settings:read"
	PermTeamSettingsWrite = "team_settings:write"

	PermMembersRead   = "members:read"
	PermMembersWrite  = "members:write"
	PermMembersDelete = "members:delete"

	PermInvitesRead   = "invites:read"
	PermInvitesWrite  = "invites:write"
	PermInvitesDelete = "invites:delete"

	PermSeasonsRead   = "seasons:read"
	PermSeasonsWrite  = "seasons:write"
	PermSeasonsDelete = "seasons:delete"

	PermGoalsRead   = "goals:read"
	PermGoalsWrite  = "goals:write"
	PermGoalsDelete = "goals:delete"

	PermProgressReportsRead   = "progress_reports:read"
	PermProgressReportsWrite  = "progress_reports:write"
	PermProgressReportsDelete = "progress_reports:delete"

	PermProgressRead  = "progress:read"
	PermProgressWrite = "progress:write"

	PermCommentsRead   = "comments:read"
	PermCommentsWrite  = "comments:write"
	PermCommentsDelete = "comments:delete"

	PermActivitiesRead = "activities:read"
)

// Resource type constants passed to CheckPermission.
const (
	ResourceTypeTeams           = "teams"
	ResourceTypeTeamSettings    = "team_settings"
	ResourceTypeMembers         = "members"
	ResourceTypeInvites         = "invites"
	ResourceTypeSeasons         = "seasons"
	ResourceTypeGoals           = "goals"
	ResourceTypeProgressReports = "progress_reports"
	ResourceTypeProgress        = "progress"
	ResourceTypeComments        = "comments"
	ResourceTypeActivities      = "activities"
)

// Resource describes the resource being accessed. OwnedBy is the direct owner
// (creator). ParentOwnedBy is the owner of the parent resource, used for
// comments where the parent goal/report owner also gets access.
type Resource struct {
	Type          string
	OwnedBy       string
	ParentOwnedBy string
}
