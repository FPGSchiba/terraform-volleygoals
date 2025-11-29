package self

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/storage"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func GetSelf(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	username := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if username == "" {
		return utils.ErrorResponse(http.StatusUnauthorized, utils.MsgErrorUnauthorized, nil)
	}
	user, err := users.GetUserBySub(ctx, username)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if user == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}
	assignments, err := db.GetTeamAssignmentsByUserID(ctx, username)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK,
		utils.MsgSuccess,
		map[string]interface{}{
			"user":        user,
			"assignments": assignments,
		})
}

func UpdateSelf(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var request UpdateSelfInput
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	username := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if username == "" {
		return utils.ErrorResponse(http.StatusUnauthorized, utils.MsgErrorUnauthorized, nil)
	}
	err = users.UpdateUserAttributes(ctx, username, request.ToCognitoAttributeMap())
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	user, err := users.GetUserBySub(ctx, username)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if user == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"user": user,
	})
}

func UploadSelfPicture(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	username := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if username == "" {
		return utils.ErrorResponse(http.StatusUnauthorized, utils.MsgErrorUnauthorized, nil)
	}
	filename, ok := event.QueryStringParameters["filename"]
	if !ok || filename == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	contentType, ok := event.QueryStringParameters["contentType"]
	if !ok || contentType == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	presignedURL, key, err := storage.GeneratePresignedUploadURLForUserPicture(ctx, username, filename, contentType, 15)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	// TODO: Update user's profile picture URL in the database if needed
	err = users.UpdateUserAttributes(ctx, username, map[string]string{
		"picture": storage.GetPublicFileURL(key),
	})
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"uploadUrl": presignedURL,
		"key":       key,
		"fileUrl":   storage.GetPublicFileURL(key),
	})
}
