package db

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

func CreateComment(ctx context.Context, authorId, commentType, targetId, content string, authorName *string, authorPicture *string) (*models.Comment, error) {
	client = GetClient()
	now := time.Now()
	comment := &models.Comment{
		Id:            models.GenerateID(),
		AuthorId:      authorId,
		AuthorName:    authorName,
		AuthorPicture: authorPicture,
		CommentType:   models.CommentType(commentType),
		TargetId:      targetId,
		Content:       content,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &commentsTableName,
		Item:      comment.ToAttributeValues(),
	})
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func GetCommentById(ctx context.Context, commentId string) (*models.Comment, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &commentsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: commentId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var comment models.Comment
	err = attributevalue.UnmarshalMap(result.Item, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func UpdateComment(ctx context.Context, commentId, content string) (*models.Comment, error) {
	client = GetClient()
	updateExpr := "SET #content = :content, #updatedAt = :updatedAt"
	result, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &commentsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: commentId},
		},
		UpdateExpression: aws.String(updateExpr),
		ExpressionAttributeNames: map[string]string{
			"#content":   "content",
			"#updatedAt": "updatedAt",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":content":   &types.AttributeValueMemberS{Value: content},
			":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		return nil, err
	}
	var updatedComment models.Comment
	err = attributevalue.UnmarshalMap(result.Attributes, &updatedComment)
	if err != nil {
		return nil, err
	}
	return &updatedComment, nil
}

func DeleteComment(ctx context.Context, commentId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &commentsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: commentId},
		},
	})
	return err
}

func ListComments(ctx context.Context, filter CommentFilter) ([]*models.Comment, int, *models.Cursor, bool, error) {
	client = GetClient()

	limit := filter.Limit
	if limit <= 0 {
		limit = defaultPageSize
	}

	in := &dynamodb.ScanInput{
		TableName: aws.String(commentsTableName),
		Limit:     aws.Int32(int32(limit)),
	}

	expr, vals, names := filter.BuildExpression()
	if strings.TrimSpace(expr) != "" {
		in.FilterExpression = aws.String(expr)
		in.ExpressionAttributeValues = vals
		if len(names) > 0 {
			in.ExpressionAttributeNames = names
		}
	}

	if filter.Cursor != nil && filter.Cursor.LastID != "" {
		in.ExclusiveStartKey = map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: filter.Cursor.LastID},
		}
	}

	result, err := client.Scan(ctx, in)
	if err != nil {
		return nil, 0, nil, false, err
	}

	comments, err := unmarshalComments(result.Items)
	if err != nil {
		return nil, 0, nil, false, err
	}

	if sortBy, sortOrder := filter.NormalizeSort(); sortBy != "" {
		sortComments(comments, sortBy, sortOrder)
	}

	nextCursor, hasMore := nextCursorFromLEK(result.LastEvaluatedKey)

	return comments, len(comments), nextCursor, hasMore, nil
}

func CreateCommentFile(ctx context.Context, commentId, storageKey string) (*models.CommentFile, error) {
	client = GetClient()
	cf := &models.CommentFile{
		Id:         models.GenerateID(),
		CommentId:  commentId,
		StorageKey: storageKey,
		CreatedAt:  time.Now(),
	}

	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &commentFilesTableName,
		Item:      cf.ToAttributeValues(),
	})
	if err != nil {
		return nil, err
	}

	return cf, nil
}

func GetCommentFilesByCommentId(ctx context.Context, commentId string) ([]*models.CommentFile, error) {
	result, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(commentFilesTableName),
		FilterExpression: aws.String("#cid = :commentId"),
		ExpressionAttributeNames: map[string]string{
			"#cid": "commentId",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":commentId": &types.AttributeValueMemberS{Value: commentId},
		},
	})
	if err != nil {
		return nil, err
	}
	files := make([]*models.CommentFile, 0, len(result.Items))
	for _, item := range result.Items {
		var cf models.CommentFile
		if err := attributevalue.UnmarshalMap(item, &cf); err != nil {
			return nil, err
		}
		files = append(files, &cf)
	}
	return files, nil
}

func unmarshalComments(items []map[string]types.AttributeValue) ([]*models.Comment, error) {
	comments := make([]*models.Comment, 0, len(items))
	for _, it := range items {
		var c models.Comment
		if err := attributevalue.UnmarshalMap(it, &c); err != nil {
			return nil, err
		}
		comments = append(comments, &c)
	}
	return comments, nil
}

func sortComments(comments []*models.Comment, sortBy, sortOrder string) {
	desc := strings.ToLower(sortOrder) == "desc"
	switch strings.ToLower(sortBy) {
	case "createdat", "created_at":
		sort.Slice(comments, func(i, j int) bool {
			if desc {
				return comments[i].CreatedAt.After(comments[j].CreatedAt)
			}
			return comments[i].CreatedAt.Before(comments[j].CreatedAt)
		})
	case "updatedat", "updated_at":
		sort.Slice(comments, func(i, j int) bool {
			if desc {
				return comments[i].UpdatedAt.After(comments[j].UpdatedAt)
			}
			return comments[i].UpdatedAt.Before(comments[j].UpdatedAt)
		})
	}
}

// listCommentsByTargetId returns all comments for a given targetId (used for cascade deletes).
func listCommentsByTargetId(ctx context.Context, targetId string) ([]*models.Comment, error) {
	result, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(commentsTableName),
		FilterExpression: aws.String("#tid = :targetId"),
		ExpressionAttributeNames: map[string]string{
			"#tid": "targetId",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":targetId": &types.AttributeValueMemberS{Value: targetId},
		},
	})
	if err != nil {
		return nil, err
	}
	return unmarshalComments(result.Items)
}

// deleteCommentFilesByCommentId deletes all comment files for a given commentId.
func deleteCommentFilesByCommentId(ctx context.Context, commentId string) error {
	result, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(commentFilesTableName),
		FilterExpression: aws.String("#cid = :commentId"),
		ExpressionAttributeNames: map[string]string{
			"#cid": "commentId",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":commentId": &types.AttributeValueMemberS{Value: commentId},
		},
	})
	if err != nil {
		return err
	}
	for _, item := range result.Items {
		var cf models.CommentFile
		if err := attributevalue.UnmarshalMap(item, &cf); err != nil {
			return err
		}
		_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: &commentFilesTableName,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: cf.Id},
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteCommentsForTarget deletes all comments (and their files) for a given targetId.
func DeleteCommentsForTarget(ctx context.Context, targetId string) error {
	comments, err := listCommentsByTargetId(ctx, targetId)
	if err != nil {
		return err
	}
	for _, c := range comments {
		if err := deleteCommentFilesByCommentId(ctx, c.Id); err != nil {
			return err
		}
		if err := DeleteComment(ctx, c.Id); err != nil {
			return err
		}
	}
	return nil
}

func GetResourceFromCommentId(ctx context.Context, commentId string) (*models.Resource, error) {
	comment, err := GetCommentById(ctx, commentId)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, nil
	}

	var resourceType string
	var parentOwnedBy string
	switch comment.CommentType {
	case models.CommentTypeProgressReport:
		resourceType = models.ResourceTypeProgressReports
		report, err := GetProgressReportById(ctx, comment.TargetId)
		if err != nil {
			return nil, err
		}
		if report == nil {
			return nil, nil
		}
		parentOwnedBy = report.AuthorId
	case models.CommentTypeGoal:
		resourceType = models.ResourceTypeGoals
		goal, err := GetGoalById(ctx, comment.TargetId)
		if err != nil {
			return nil, err
		}
		if goal == nil {
			return nil, nil
		}
		parentOwnedBy = goal.OwnerId
	case models.CommentTypeProgressEntry:
		resourceType = models.ResourceTypeProgress
		progress, err := GetProgressById(ctx, comment.TargetId)
		if err != nil {
			return nil, err
		}
		if progress == nil {
			return nil, nil
		}
		report, err := GetProgressReportById(ctx, progress.ProgressReportId)
		if err != nil {
			return nil, err
		}
		if report == nil {
			return nil, nil
		}
		parentOwnedBy = report.AuthorId
	default:
		return nil, nil
	}

	return &models.Resource{
		Type:          resourceType,
		OwnedBy:       comment.AuthorId,
		ParentOwnedBy: parentOwnedBy,
	}, nil
}
