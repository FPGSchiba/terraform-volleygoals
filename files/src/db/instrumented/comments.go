package instrumented

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func CreateComment(ctx context.Context, teamId, actorId, commentType, targetId, content string, authorName, authorPicture *string, targetOwnerId string) (*models.Comment, error) {
	comment, err := db.CreateComment(ctx, actorId, commentType, targetId, content, authorName, authorPicture)
	if err != nil {
		return nil, err
	}
	activity.EmitCommentCreated(ctx, teamId, actorId, comment.Id, targetOwnerId)
	return comment, nil
}

func UpdateComment(ctx context.Context, teamId, actorId, commentId, content, targetOwnerId string) (*models.Comment, error) {
	comment, err := db.UpdateComment(ctx, commentId, content)
	if err != nil {
		return nil, err
	}
	activity.EmitCommentUpdated(ctx, teamId, actorId, commentId, targetOwnerId)
	return comment, nil
}

func DeleteComment(ctx context.Context, teamId, actorId, commentId, targetOwnerId string) error {
	if err := db.DeleteComment(ctx, commentId); err != nil {
		return err
	}
	activity.EmitCommentDeleted(ctx, teamId, actorId, commentId, targetOwnerId)
	return nil
}
