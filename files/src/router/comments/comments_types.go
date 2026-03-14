package comments

import "github.com/fpgschiba/volleygoals/models"

type CreateCommentRequest struct {
	CommentType models.CommentType `json:"commentType"`
	TargetId    string             `json:"targetId"`
	Content     string             `json:"content"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}
