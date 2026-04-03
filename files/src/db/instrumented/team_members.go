package instrumented

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func AddTeamMember(ctx context.Context, teamId, actorId, userId string, role models.TeamMemberRole) (*models.TeamMember, error) {
	member, err := db.AddTeamMember(ctx, teamId, userId, role)
	if err != nil {
		return nil, err
	}
	activity.EmitMemberJoined(ctx, teamId, userId)
	return member, nil
}

func UpdateTeamMemberRole(ctx context.Context, teamId, actorId, memberId string, role models.TeamMemberRole) (*models.TeamMember, error) {
	member, err := db.UpdateTeamMember(ctx, memberId, &role, nil)
	if err != nil {
		return nil, err
	}
	activity.EmitMemberRoleChanged(ctx, teamId, actorId, role, memberId)
	return member, nil
}

func RemoveTeamMember(ctx context.Context, teamId, actorId, memberId string) error {
	if err := db.RemoveTeamMember(ctx, memberId); err != nil {
		return err
	}
	activity.EmitMemberRemoved(ctx, teamId, actorId, memberId)
	return nil
}
