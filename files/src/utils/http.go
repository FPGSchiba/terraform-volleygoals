package utils

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

type ResponseMessage string

const (
	// General errors
	MsgGeneralError        ResponseMessage = "error.general"
	MsgInternalServerError ResponseMessage = "error.internalServerError"
	MsgBadRequest          ResponseMessage = "error.badRequest"
	MsgErrorForbidden      ResponseMessage = "error.forbidden"
	MsgErrorNotFound       ResponseMessage = "error.notFound"
	MsgNotImplemented      ResponseMessage = "error.notImplemented"
	MsgErrorUnauthorized   ResponseMessage = "error.unauthorized"

	// Team related errors
	MsgErrorTeamExists   ResponseMessage = "error.team.exists"
	MsgErrorTeamNotFound ResponseMessage = "error.team.notFound"

	// Team Settings related errors
	MsgErrorTeamSettingsNotFound ResponseMessage = "error.teamSettings.notFound"

	// Invite related errors
	MsgErrorInviteExists           ResponseMessage = "error.invite.exists"
	MsgErrorInvalidInviteToken     ResponseMessage = "error.invite.invalidToken"
	MsgErrorInviteExpired          ResponseMessage = "error.invite.expired"
	MsgErrorInviteAlreadyCompleted ResponseMessage = "error.invite.alreadyCompleted"
	MsgErrorUserAlreadyMember      ResponseMessage = "error.invite.userAlreadyMember"

	// General Success messages
	MsgSuccess ResponseMessage = "success.ok"

	// Team related success messages
	MsgSuccessTeamCreated ResponseMessage = "success.team.created"
	MsgSuccessTeamDeleted ResponseMessage = "success.team.deleted"
)

func Response(status int, body interface{}) (*events.APIGatewayProxyResponse, error) {
	resp := events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Content-Type":                     "application/json",
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Headers":     "Content-Type",
			"Access-Control-Allow-Methods":     "OPTIONS, POST, GET, PUT, DELETE",
			"Access-Control-Allow-Credentials": "true",
		},
	}
	resp.StatusCode = status

	// Convert body to json data
	sBody, _ := json.Marshal(body)
	resp.Body = string(sBody)

	return &resp, nil
}

func SuccessResponse(status int, message ResponseMessage, data interface{}) (*events.APIGatewayProxyResponse, error) {
	respMap := map[string]interface{}{
		"message": string(message),
	}
	if data == nil {
		return Response(status, respMap)
	}
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			respMap[k] = v
		}
		return Response(status, respMap)
	}
	// Fallback: include data under "data" key
	respMap["data"] = data
	return Response(status, respMap)
}

func ErrorResponse(status int, message ResponseMessage, err error) (*events.APIGatewayProxyResponse, error) {
	if err == nil {
		return Response(status, map[string]interface{}{"message": string(message)})
	}
	return Response(status, map[string]interface{}{"message": string(message), "error": err.Error()})
}
