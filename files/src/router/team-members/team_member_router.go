package team_members

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/utils"
)

func ListTeamMembers(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement listing team members
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func AddTeamMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement adding a team member
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func UpdateTeamMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement updating a team member
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func RemoveTeamMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement removing a team member
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func LeaveTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	// TODO: implement leaving a team
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}
