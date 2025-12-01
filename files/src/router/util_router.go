package router

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/utils"
)

func HealthCheck(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement actually functioning health check which checks dependencies (AWS Health? DB connectivity?)
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}
