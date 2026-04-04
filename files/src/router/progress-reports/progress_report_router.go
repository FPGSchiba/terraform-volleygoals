package progress_reports

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/db/instrumented"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func CreateProgressReport(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	if seasonId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if teamId == "" {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeProgressReports}, models.PermProgressReportsWrite) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	var request CreateProgressReportRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	authorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)

	user, uerr := users.GetUserBySub(ctx, authorId)
	if uerr != nil {
		log.Printf("CreateProgressReport: failed to fetch user %s: %v", authorId, uerr)
	}
	var authorName *string
	if user != nil {
		switch {
		case user.Name != nil && *user.Name != "":
			authorName = user.Name
		case user.PreferredUsername != nil && *user.PreferredUsername != "":
			authorName = user.PreferredUsername
		default:
			authorName = aws.String(user.Email)
		}
	}
	var authorPicture *string
	if user != nil {
		authorPicture = user.Picture
	}

	var entries []db.ProgressEntry
	for _, p := range request.Progress {
		entries = append(entries, db.ProgressEntry{GoalId: p.GoalId, Rating: p.Rating, Details: p.Details})
	}

	report, err := db.CreateProgressReport(ctx, seasonId, authorId, request.Summary, request.Details, request.OverallDetails, request.ReportedAt, entries, authorName, authorPicture)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	activity.EmitProgressReportCreated(ctx, teamId, authorId, report.Id, report.AuthorId)

	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{
		"progressReport": report,
	})
}

func GetProgressReport(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	reportId := event.PathParameters["reportId"]
	if seasonId == "" || reportId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if teamId == "" {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	report, err := db.GetProgressReportById(ctx, reportId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if report == nil || report.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorProgressReportNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	allowed, err := utils.CheckPermission(ctx, actorId, teamId,
		models.Resource{Type: models.ResourceTypeProgressReports, OwnedBy: report.AuthorId},
		models.PermProgressReportsRead)
	if err != nil || !allowed {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	// Enrich AuthorName/AuthorPicture for legacy records (same logic as ListProgressReports).
	if report.AuthorName == nil {
		user, uerr := users.GetUserBySub(ctx, report.AuthorId)
		if uerr != nil {
			log.Printf("GetProgressReport: failed to fetch user %s: %v", report.AuthorId, uerr)
		}
		if user != nil {
			switch {
			case user.Name != nil && *user.Name != "":
				report.AuthorName = user.Name
			case user.PreferredUsername != nil && *user.PreferredUsername != "":
				report.AuthorName = user.PreferredUsername
			default:
				name := user.Email
				report.AuthorName = &name
			}
			if report.AuthorPicture == nil {
				report.AuthorPicture = user.Picture
			}
		}
	}

	entries := make([]*models.Progress, 0)
	fetchedEntries, err := db.ListProgressEntriesByReportId(ctx, reportId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if fetchedEntries != nil {
		entries = fetchedEntries
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"progressReport": ProgressReportWithProgress{report, entries},
	})
}

func ListProgressReports(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	if seasonId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if teamId == "" {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeProgressReports}, models.PermProgressReportsRead) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	filter, err := db.ProgressReportFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	filter.SeasonId = seasonId

	items, count, nextCursor, hasMore, err := db.ListProgressReports(ctx, filter)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	reportIds := make([]string, 0, len(items))
	for _, r := range items {
		reportIds = append(reportIds, r.Id)
	}
	progressByReport, err := db.ListProgressEntriesByReportIds(ctx, reportIds)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	// Resolve authorName/authorPicture for reports missing them (legacy records).
	uniqueAuthorIds := make(map[string]struct{})
	for _, r := range items {
		if r.AuthorName == nil {
			uniqueAuthorIds[r.AuthorId] = struct{}{}
		}
	}
	userByAuthorId := make(map[string]*models.User)
	for authorId := range uniqueAuthorIds {
		u, uerr := users.GetUserBySub(ctx, authorId)
		if uerr != nil {
			log.Printf("ListProgressReports: failed to fetch user %s: %v", authorId, uerr)
			continue
		}
		if u != nil {
			userByAuthorId[authorId] = u
		}
	}
	for _, r := range items {
		if r.AuthorName != nil {
			continue
		}
		u, ok := userByAuthorId[r.AuthorId]
		if !ok {
			continue
		}
		switch {
		case u.Name != nil && *u.Name != "":
			r.AuthorName = u.Name
		case u.PreferredUsername != nil && *u.PreferredUsername != "":
			r.AuthorName = u.PreferredUsername
		default:
			name := u.Email
			r.AuthorName = &name
		}
		if r.AuthorPicture == nil {
			r.AuthorPicture = u.Picture
		}
	}

	enrichedItems := make([]ProgressReportWithProgress, 0, len(items))
	for _, r := range items {
		entries := progressByReport[r.Id]
		if entries == nil {
			entries = make([]*models.Progress, 0)
		}
		enrichedItems = append(enrichedItems, ProgressReportWithProgress{r, entries})
	}

	nextToken := ""
	if nextCursor != nil {
		nextToken, err = models.EncodeCursor(nextCursor)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items":     enrichedItems,
		"count":     count,
		"nextToken": nextToken,
		"hasMore":   hasMore,
	})
}

func UpdateProgressReport(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	reportId := event.PathParameters["reportId"]
	if seasonId == "" || reportId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if teamId == "" {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	report, err := db.GetProgressReportById(ctx, reportId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if report == nil || report.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorProgressReportNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	allowed, err := utils.CheckPermission(ctx, actorId, teamId,
		models.Resource{Type: models.ResourceTypeProgressReports, OwnedBy: report.AuthorId},
		models.PermProgressReportsWrite)
	if err != nil || !allowed {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	var request UpdateProgressReportRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	var entries []db.ProgressEntry
	for _, p := range request.Progress {
		entries = append(entries, db.ProgressEntry{GoalId: p.GoalId, Rating: p.Rating, Details: p.Details})
	}

	updatedReport, err := instrumented.UpdateProgressReport(ctx, teamId, actorId, reportId, request.Summary, request.Details, request.OverallDetails, entries, report.AuthorId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"progressReport": updatedReport,
	})
}

func DeleteProgressReport(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	reportId := event.PathParameters["reportId"]
	if seasonId == "" || reportId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	teamId, err := db.GetTeamIdBySeasonId(ctx, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if teamId == "" {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	report, err := db.GetProgressReportById(ctx, reportId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if report == nil || report.SeasonId != seasonId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorProgressReportNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	allowed, err := utils.CheckPermission(ctx, actorId, teamId,
		models.Resource{Type: models.ResourceTypeProgressReports, OwnedBy: report.AuthorId},
		models.PermProgressReportsDelete)
	if err != nil || !allowed {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	if err := instrumented.DeleteProgressReport(ctx, teamId, actorId, reportId, report.AuthorId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusNoContent, utils.MsgSuccess, nil)
}
