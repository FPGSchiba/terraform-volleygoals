package router

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/utils"
)

func HealthCheck(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement uploading team picture
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func GetFile(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement getting a file
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}
