package progress_reports

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/utils"
)

func CreateProgressReport(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to create a comment
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func GetProgressReport(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to get a comment
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func ListProgressReports(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to list comments
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func UpdateProgressReport(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to update a comment
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func DeleteProgressReport(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: Implement the logic to delete a comment
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}
