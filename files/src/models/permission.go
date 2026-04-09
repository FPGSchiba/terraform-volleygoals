package models

// TenantIDGlobal is the sentinel tenantId value for global default policies.
// DynamoDB cannot index null values, so "global" is used instead of an empty string.
const TenantIDGlobal = "global"

// Permission variables — resource:action pairs used in RoleDefinition.Permissions
// and OwnershipPolicy.OwnerPermissions. Values are populated at init from
// the canonical resource definitions in models.GetDefinitions().
var (
	PermTeamsRead   string
	PermTeamsWrite  string
	PermTeamsDelete string

	PermTeamSettingsRead  string
	PermTeamSettingsWrite string

	PermMembersRead   string
	PermMembersWrite  string
	PermMembersDelete string

	PermInvitesRead   string
	PermInvitesWrite  string
	PermInvitesDelete string

	PermSeasonsRead   string
	PermSeasonsWrite  string
	PermSeasonsDelete string

	PermTeamGoalsRead   string
	PermTeamGoalsWrite  string
	PermTeamGoalsDelete string

	PermIndividualGoalsRead   string
	PermIndividualGoalsWrite  string
	PermIndividualGoalsDelete string

	PermProgressReportsRead   string
	PermProgressReportsWrite  string
	PermProgressReportsDelete string

	PermProgressRead  string
	PermProgressWrite string

	PermCommentsRead   string
	PermCommentsWrite  string
	PermCommentsDelete string

	PermActivitiesRead string
)

func init() {
	// Populate variables with generated values using the canonical accessor.
	// Fall back to the original hard-coded strings if the model does not
	// contain the resource/action (GetPermission returns a best-effort
	// formatted string in that case).
	PermTeamsRead = GetPermission("teams", "read")
	PermTeamsWrite = GetPermission("teams", "write")
	PermTeamsDelete = GetPermission("teams", "delete")

	PermTeamSettingsRead = GetPermission("team_settings", "read")
	PermTeamSettingsWrite = GetPermission("team_settings", "write")

	PermMembersRead = GetPermission("members", "read")
	PermMembersWrite = GetPermission("members", "write")
	PermMembersDelete = GetPermission("members", "delete")

	PermInvitesRead = GetPermission("invites", "read")
	PermInvitesWrite = GetPermission("invites", "write")
	PermInvitesDelete = GetPermission("invites", "delete")

	PermSeasonsRead = GetPermission("seasons", "read")
	PermSeasonsWrite = GetPermission("seasons", "write")
	PermSeasonsDelete = GetPermission("seasons", "delete")

	PermTeamGoalsRead = GetPermission("team_goals", "read")
	PermTeamGoalsWrite = GetPermission("team_goals", "write")
	PermTeamGoalsDelete = GetPermission("team_goals", "delete")

	PermIndividualGoalsRead = GetPermission("individual_goals", "read")
	PermIndividualGoalsWrite = GetPermission("individual_goals", "write")
	PermIndividualGoalsDelete = GetPermission("individual_goals", "delete")

	PermProgressReportsRead = GetPermission("progress_reports", "read")
	PermProgressReportsWrite = GetPermission("progress_reports", "write")
	PermProgressReportsDelete = GetPermission("progress_reports", "delete")

	PermProgressRead = GetPermission("progress", "read")
	PermProgressWrite = GetPermission("progress", "write")

	PermCommentsRead = GetPermission("comments", "read")
	PermCommentsWrite = GetPermission("comments", "write")
	PermCommentsDelete = GetPermission("comments", "delete")

	PermActivitiesRead = GetPermission("activities", "read")
}

// Resource type constants passed to CheckPermission.
const (
	ResourceTypeTeams           = "teams"
	ResourceTypeTeamSettings    = "team_settings"
	ResourceTypeMembers         = "members"
	ResourceTypeInvites         = "invites"
	ResourceTypeSeasons         = "seasons"
	ResourceTypeTeamGoals       = "team_goals"
	ResourceTypeIndividualGoals = "individual_goals"
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
