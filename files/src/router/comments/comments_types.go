package comments

import "github.com/fpgschiba/volleygoals/models"

type CommentFileResult struct {
	Id         string `json:"id"`
	CommentId  string `json:"commentId"`
	StorageKey string `json:"storageKey"`
	FileUrl    string `json:"fileUrl"`
}

type CreateCommentRequest struct {
	CommentType models.CommentType `json:"commentType"`
	TargetId    string             `json:"targetId"`
	Content     string             `json:"content"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}
