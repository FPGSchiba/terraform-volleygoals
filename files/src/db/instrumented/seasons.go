package instrumented

import (
	"context"
	"time"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func CreateSeason(ctx context.Context, teamId, actorId, name string, start, end time.Time) (*models.Season, error) {
	season, err := db.CreateSeason(ctx, teamId, name, start, end)
	if err != nil {
		return nil, err
	}
	activity.EmitSeasonCreated(ctx, teamId, actorId, season.Name, season.Id)
	return season, nil
}

func UpdateSeason(ctx context.Context, teamId, actorId, seasonId string, name *string, start, end *time.Time, status *models.SeasonStatus) (*models.Season, error) {
	season, err := db.UpdateSeason(ctx, seasonId, name, start, end, status)
	if err != nil {
		return nil, err
	}
	activity.EmitSeasonUpdated(ctx, teamId, actorId, season.Name, seasonId)
	return season, nil
}

func DeleteSeason(ctx context.Context, teamId, actorId, seasonId, seasonName string) error {
	if err := db.DeleteSeason(ctx, seasonId); err != nil {
		return err
	}
	activity.EmitSeasonDeleted(ctx, teamId, actorId, seasonName, seasonId)
	return nil
}
