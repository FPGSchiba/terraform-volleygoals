package router

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/utils"
	log "github.com/sirupsen/logrus"
)

func GetSelf(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	username := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if username == "" {
		return utils.ErrorResponse(http.StatusUnauthorized, utils.MsgErrorUnauthorized, nil)
	}
	log.WithField("username", username).Info("GetSelf called")
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func UpdateSelf(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}
