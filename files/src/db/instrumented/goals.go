package instrumented

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func CreateGoal(ctx context.Context, teamId, actorId string, goalType models.GoalType, title, description string) (*models.Goal, error) {
	goal, err := db.CreateGoal(ctx, teamId, actorId, goalType, title, description)
	if err != nil {
		return nil, err
	}
	activity.EmitGoalCreated(ctx, teamId, actorId, goal.Title, goal.Id, goal.OwnerId)
	return goal, nil
}

func DeleteGoal(ctx context.Context, teamId, actorId, goalId, goalTitle, goalOwnerId string) error {
	if err := db.DeleteGoal(ctx, goalId); err != nil {
		return err
	}
	activity.EmitGoalDeleted(ctx, teamId, actorId, goalTitle, goalId, goalOwnerId)
	return nil
}
