package router

import (
	"context"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/utils"
)

func GetSelf(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	username := utils.GetUserCognitoUsername(event.RequestContext.Authorizer)
	if username == "" {
		return utils.ErrorResponse(http.StatusUnauthorized, utils.MsgErrorUnauthorized, nil)
	}
	log.Printf("GetSelf: username: %s", username)
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func UpdateSelf(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}
