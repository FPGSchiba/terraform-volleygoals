package instrumented

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func UpdateProgressReport(ctx context.Context, teamId, actorId, reportId string, summary, details, overallDetails *string, entries []db.ProgressEntry, ownerId string) (*models.ProgressReport, error) {
	report, err := db.UpdateProgressReport(ctx, reportId, summary, details, overallDetails, entries)
	if err != nil {
		return nil, err
	}
	activity.EmitProgressReportUpdated(ctx, teamId, actorId, reportId, ownerId)
	return report, nil
}

func DeleteProgressReport(ctx context.Context, teamId, actorId, reportId, ownerId string) error {
	if err := db.DeleteProgressReport(ctx, reportId); err != nil {
		return err
	}
	activity.EmitProgressReportDeleted(ctx, teamId, actorId, reportId, ownerId)
	return nil
}
