package invites

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/mail"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func CompleteInvite(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	req, resp, err := parseCompleteInviteRequest(event.Body)
	if resp != nil && err != nil {
		return resp, err
	}

	invite, resp, err := fetchAndValidateInvite(ctx, req.Token)
	if resp != nil {
		return resp, err
	}

	userSub, created, resp, err := getOrCreateUserForInvite(ctx, req, invite)
	if resp != nil {
		return resp, err
	}

	invite, member, resp, err := finalizeInvite(ctx, invite, userSub, req.Accepted, created)
	if resp != nil {
		return resp, err
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"invite": invite,
		"member": member,
	})
}

func CreateInvite(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var request CreateInviteRequest
	err := json.Unmarshal([]byte(event.Body), &request)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasOneRoleOnTeam(ctx, event.RequestContext.Authorizer, request.TeamId, []models.TeamMemberRole{models.TeamMemberRoleAdmin, models.TeamMemberRoleTrainer}) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	inviteToken := utils.GenerateInviteToken(request.TeamId, request.Email, request.Role)
	exists, err := db.DoesInviteExistByToken(ctx, inviteToken)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if exists {
		return utils.SuccessResponse(http.StatusConflict, utils.MsgErrorInviteExists, map[string]interface{}{"inviteToken": inviteToken})
	}
	currentUsername := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	invite, err := db.CreateInvite(ctx, request.TeamId, request.Email, currentUsername, inviteToken, request.Role)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	currentUser, err := users.GetUserBySub(ctx, currentUsername)
	if err != nil || currentUser == nil {
		_ = db.RemoveInviteById(ctx, invite.Id)
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	var inviterName string
	if currentUser.Name == nil || *currentUser.Name == "" {
		inviterName = currentUser.Email
	} else {
		inviterName = *currentUser.Name
	}
	team, err := db.GetTeamById(ctx, request.TeamId)
	if err != nil || team == nil {
		_ = db.RemoveInviteById(ctx, invite.Id)
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	err = mail.SendInvitationEmail(ctx, request.Email, inviteToken, team.Name, inviterName, db.InviteExpiresInDays)
	if err != nil {
		_ = db.RemoveInviteById(ctx, invite.Id)
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"invite": invite,
	})
}

func ListInvites(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func RevokeInvite(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func ResendInvite(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return utils.ErrorResponse(http.StatusNotImplemented, utils.MsgNotImplemented, nil)
}

func parseCompleteInviteRequest(body string) (CompleteInviteRequest, *events.APIGatewayProxyResponse, error) {
	var req CompleteInviteRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		resp, e := utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
		return req, resp, e
	}
	return req, nil, nil
}

func fetchAndValidateInvite(ctx context.Context, token string) (*models.Invite, *events.APIGatewayProxyResponse, error) {
	invite, err := db.GetInviteByToken(ctx, token)
	if err != nil {
		resp, e := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return nil, resp, e
	}
	if invite == nil {
		resp, e := utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
		return nil, resp, e
	}
	if !utils.ValidateInviteToken(token, invite.Email, invite.TeamId, invite.Role) {
		resp, e := utils.ErrorResponse(http.StatusBadRequest, utils.MsgErrorInvalidInviteToken, nil)
		return nil, resp, e
	}
	return invite, nil, nil
}

func getOrCreateUserForInvite(ctx context.Context, req CompleteInviteRequest, invite *models.Invite) (string, bool, *events.APIGatewayProxyResponse, error) {
	existingUser, err := users.GetUserByEmail(ctx, req.Email)
	if err != nil {
		resp, e := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return "", false, resp, e
	}
	if existingUser != nil {
		return existingUser.Id, false, nil, nil
	}

	// No existing user
	if !req.Accepted {
		if _, err := db.CompleteInvite(ctx, invite.Id, "", false); err != nil {
			resp, e := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
			return "", false, resp, e
		}
		resp, e := utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"invite": invite})
		return "", false, resp, e
	}

	createdUser, err := users.CreateUser(ctx, req.Email)
	if err != nil {
		resp, e := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return "", false, resp, e
	}
	return createdUser.Id, true, nil, nil
}

func finalizeInvite(ctx context.Context, invite *models.Invite, userSub string, accepted bool, created bool) (*models.Invite, *models.TeamMember, *events.APIGatewayProxyResponse, error) {
	invite, err := db.CompleteInvite(ctx, invite.Id, userSub, accepted)
	if err != nil {
		if created && userSub != "" {
			_ = users.DeleteUserBySub(ctx, userSub)
		}
		resp, e := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return nil, nil, resp, e
	}

	member, err := db.CreateTeamMemberFromInvite(ctx, invite)
	if err != nil {
		if created && userSub != "" {
			_ = users.DeleteUserBySub(ctx, userSub)
		}
		resp, e := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return nil, nil, resp, e
	}

	return invite, member, nil, nil
}
