package invites

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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
	if invite == nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgErrorInviteAlreadyCompleted, nil)
	}

	if invite.ExpiresAt.Before(time.Now()) {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgErrorInviteExpired, nil)
	}

	userSub, tempPassword, created, resp, err := getOrCreateUserForInvite(ctx, req, invite)
	if resp != nil {
		return resp, err
	}

	invite, member, resp, err := finalizeInvite(ctx, invite, userSub, req.Accepted, created)
	if resp != nil {
		return resp, err
	}

	result := map[string]interface{}{
		"invite":      invite,
		"member":      member,
		"userCreated": created,
	}
	if tempPassword != "" {
		result["temporaryPassword"] = tempPassword
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, result)
}

func CreateInvite(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var request CreateInviteRequest
	req, resp, err := parseCreateInviteRequest(event.Body)
	if resp != nil && err != nil {
		return resp, err
	}
	request = req

	// Authorization
	if resp, err := authorizeCreateInvite(ctx, event.RequestContext.Authorizer, request.TeamId); resp != nil {
		return resp, err
	}

	// Token and existence check
	inviteToken := utils.GenerateInviteToken(request.TeamId, request.Email, request.Role)
	if resp, err := ensureInviteTokenAvailable(ctx, inviteToken); resp != nil {
		return resp, err
	}

	// Existing user / membership checks
	if resp, err := checkExistingUserMembership(ctx, request.Email, request.TeamId); resp != nil {
		return resp, err
	}

	// Create invite and send email
	currentUsername := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	invite, resp, err := createInviteAndNotify(ctx, request, currentUsername, inviteToken, request.SendEmail)
	if resp != nil {
		return resp, err
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"invite": invite,
	})
}

func GetTeamInvites(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	// Authorization
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasOneRoleOnTeam(ctx, event.RequestContext.Authorizer, teamId, []models.TeamMemberRole{models.TeamMemberRoleAdmin, models.TeamMemberRoleTrainer}) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	filter, err := db.TeamInviteFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}

	invites, count, nextCursor, hasMore, err := db.GetInvitesByTeamId(ctx, teamId, filter)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	nextToken := ""
	if nextCursor != nil {
		nextToken, err = models.EncodeCursor(nextCursor)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
	}

	resp := models.PaginationResponse{
		Items:     invites,
		Count:     count,
		NextToken: nextToken,
		HasMore:   hasMore,
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items":     resp.Items,
		"count":     resp.Count,
		"nextToken": resp.NextToken,
		"hasMore":   resp.HasMore,
	})
}

func RevokeInvite(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	inviteId, ok := event.PathParameters["inviteId"]
	if !ok || inviteId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	invite, err := db.GetInviteById(ctx, inviteId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if invite == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, invite.TeamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	username := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	invite, err = db.RevokeInviteById(ctx, inviteId, username)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"invite": invite,
	})
}

func ResendInvite(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	inviteId, ok := event.PathParameters["inviteId"]
	if !ok || inviteId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	invite, err := db.GetInviteById(ctx, inviteId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if invite == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, invite.TeamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	inviter, err := users.GetUserBySub(ctx, invite.InvitedBy)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	team, err := db.GetTeamById(ctx, invite.TeamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	err = db.ResentInviteEmail(ctx, inviteId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	err = mail.ResendInvitationEmail(ctx, invite, inviter, team)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"invite": invite,
	})
}

func GetInviteByToken(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	token, ok := event.PathParameters["token"]
	if !ok || token == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	invite, resp, err := fetchAndValidateInvite(ctx, token)
	if resp != nil {
		return resp, err
	}
	if invite == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}
	if invite.AcceptedBy != nil {
		teamMember, err := db.GetTeamMemberByUserIDAndTeamID(ctx, *invite.AcceptedBy, invite.TeamId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
			"invite": invite,
			"member": teamMember,
		})
	}
	return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
}

func parseCreateInviteRequest(body string) (CreateInviteRequest, *events.APIGatewayProxyResponse, error) {
	var req CreateInviteRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		resp, e := utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
		return req, resp, e
	}
	return req, nil, nil
}

func authorizeCreateInvite(ctx context.Context, authorizer map[string]interface{}, teamId string) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(authorizer) && !utils.HasOneRoleOnTeam(ctx, authorizer, teamId, []models.TeamMemberRole{models.TeamMemberRoleAdmin, models.TeamMemberRoleTrainer}) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	return nil, nil
}

func ensureInviteTokenAvailable(ctx context.Context, token string) (*events.APIGatewayProxyResponse, error) {
	exists, err := db.DoesInviteExistByToken(ctx, token)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if exists {
		return utils.SuccessResponse(http.StatusConflict, utils.MsgErrorInviteExists, map[string]interface{}{"inviteToken": token})
	}
	return nil, nil
}

func checkExistingUserMembership(ctx context.Context, email, teamId string) (*events.APIGatewayProxyResponse, error) {
	existingUser, err := users.GetUserByEmail(ctx, email)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if existingUser != nil {
		isMember, err := db.IsUserMemberOfTeam(ctx, existingUser.Id, teamId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if isMember {
			return utils.ErrorResponse(http.StatusConflict, utils.MsgErrorUserAlreadyMember, nil)
		}
	}
	return nil, nil
}

func createInviteAndNotify(ctx context.Context, request CreateInviteRequest, currentUsername, inviteToken string, sendEmail bool) (*models.Invite, *events.APIGatewayProxyResponse, error) {
	invite, err := db.CreateInvite(ctx, request.TeamId, request.Email, currentUsername, inviteToken, request.Role, request.Message)
	if err != nil {
		resp, err := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return nil, resp, err
	}

	currentUser, err := users.GetUserBySub(ctx, currentUsername)
	if err != nil || currentUser == nil {
		_ = db.RemoveInviteById(ctx, invite.Id)
		resp, err := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return nil, resp, err
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
		resp, err := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return nil, resp, err
	}

	if !sendEmail {
		return invite, nil, nil
	}

	var message string
	if request.Message != nil {
		message = *request.Message
	} else {
		message = "Welcome to the team!"
	}

	// Send invitation email
	err = mail.SendInvitationEmail(ctx, request.Email, inviteToken, team.Name, inviterName, message, db.InviteExpiresInDays)
	if err != nil {
		_ = db.RemoveInviteById(ctx, invite.Id)
		resp, err := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return nil, resp, err
	}

	return invite, nil, nil
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
	if invite.ExpiresAt.Before(time.Now()) {
		resp, e := utils.ErrorResponse(http.StatusBadRequest, utils.MsgErrorInviteExpired, nil)
		_, err := db.ExpireInviteById(ctx, invite.Id)
		if err != nil {
			return nil, resp, err
		}
		return nil, resp, e
	}
	
	if !utils.ValidateInviteToken(token, invite.Email, invite.TeamId, invite.Role) {
		resp, e := utils.ErrorResponse(http.StatusBadRequest, utils.MsgErrorInvalidInviteToken, nil)
		return nil, resp, e
	}
	return invite, nil, nil
}

func getOrCreateUserForInvite(ctx context.Context, req CompleteInviteRequest, invite *models.Invite) (string, string, bool, *events.APIGatewayProxyResponse, error) {
	existingUser, err := users.GetUserByEmail(ctx, req.Email)
	if err != nil {
		resp, e := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return "", "", false, resp, e
	}
	if existingUser != nil {
		return existingUser.Id, "", false, nil, nil
	}

	// No existing user
	if !req.Accepted {
		if _, err := db.CompleteInvite(ctx, invite.Id, "", false); err != nil {
			resp, e := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
			return "", "", false, resp, e
		}
		resp, e := utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"invite": invite})
		return "", "", false, resp, e
	}

	createdUser, tempPassword, err := users.CreateUser(ctx, req.Email)
	if err != nil {
		resp, e := utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		return "", "", false, resp, e
	}
	return createdUser.Id, tempPassword, true, nil, nil
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
