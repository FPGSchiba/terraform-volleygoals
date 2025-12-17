package comments

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/utils"
)

func CreateComment(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to create a comment
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func GetComment(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to get a comment
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func ListComments(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to list comments
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func UpdateComment(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to update a comment
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func DeleteComment(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to delete a comment
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func UploadCommentFile(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to upload a comment file
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}
