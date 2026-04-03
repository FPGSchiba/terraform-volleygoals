package utils_test

import (
	"testing"

	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
	"github.com/stretchr/testify/assert"
)

func preloadedData(memberRole string, ownerPerms, rolePerms []string, resourceType string) *utils.PreloadedData {
	return &utils.PreloadedData{
		Member: &models.TeamMember{Role: models.TeamMemberRole(memberRole)},
		Team:   &models.Team{Id: "team-1"},
		OwnershipByType: map[string]*models.OwnershipPolicy{
			resourceType: {OwnerPermissions: ownerPerms},
		},
		RoleGlobal: &models.RoleDefinition{Permissions: rolePerms},
	}
}

func TestCanReadActivity_OwnerCanReadOwnGoalActivity(t *testing.T) {
	pd := preloadedData("member", []string{models.PermGoalsRead}, nil, models.ResourceTypeGoals)
	a := &models.Activity{TargetType: models.ResourceTypeGoals, TargetOwnerId: "user-1"}
	assert.True(t, pd.CanReadActivity("user-1", a))
}

func TestCanReadActivity_NonOwnerMemberCannotReadGoalActivity(t *testing.T) {
	pd := preloadedData("member", []string{models.PermGoalsRead}, nil, models.ResourceTypeGoals)
	a := &models.Activity{TargetType: models.ResourceTypeGoals, TargetOwnerId: "user-2"}
	assert.False(t, pd.CanReadActivity("user-1", a))
}

func TestCanReadActivity_TrainerCanReadAnyGoalActivity(t *testing.T) {
	pd := preloadedData("trainer", nil, []string{models.PermGoalsRead}, models.ResourceTypeGoals)
	a := &models.Activity{TargetType: models.ResourceTypeGoals, TargetOwnerId: "user-2"}
	assert.True(t, pd.CanReadActivity("trainer-1", a))
}

func TestCanReadActivity_NilMemberDenies(t *testing.T) {
	pd := &utils.PreloadedData{}
	a := &models.Activity{TargetType: models.ResourceTypeGoals, TargetOwnerId: "user-1"}
	assert.False(t, pd.CanReadActivity("user-1", a))
}

func TestCanReadActivity_SeasonActivityVisibleToAllMembers(t *testing.T) {
	pd := preloadedData("member", nil, []string{models.PermSeasonsRead}, models.ResourceTypeSeasons)
	a := &models.Activity{TargetType: models.ResourceTypeSeasons, TargetOwnerId: ""}
	assert.True(t, pd.CanReadActivity("member-1", a))
}
