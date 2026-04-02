package search

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

const (
	defaultSearchLimit = 10
	maxSearchLimit     = 50
)

func GlobalSearch(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	q := event.QueryStringParameters

	query := strings.TrimSpace(q["q"])
	if query == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	teamId := strings.TrimSpace(q["teamId"])
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	limit := defaultSearchLimit
	if v, ok := q["limit"]; ok {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	// Parse types filter; empty means search all.
	searchGoals := true
	searchReports := true
	if typesParam := strings.TrimSpace(q["types"]); typesParam != "" {
		parts := strings.Split(typesParam, ",")
		searchGoals = false
		searchReports = false
		for _, p := range parts {
			switch strings.TrimSpace(p) {
			case "goals":
				searchGoals = true
			case "reports":
				searchReports = true
			}
		}
	}

	results := make([]interface{}, 0)

	if searchGoals {
		goals, err := db.SearchGoalsForTeam(ctx, teamId, query, limit)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		for _, g := range goals {
			results = append(results, SearchGoalResult{
				Type:   "goal",
				Id:     g.Id,
				Title:  g.Title,
				TeamId: g.TeamId,
				Status: string(g.Status),
			})
		}
	}

	if searchReports && len(results) < limit {
		reportLimit := limit - len(results)
		reports, err := db.SearchProgressReportsForTeam(ctx, teamId, query, reportLimit)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		for _, r := range reports {
			results = append(results, SearchReportResult{
				Type:      "report",
				Id:        r.Id,
				Summary:   r.Summary,
				SeasonId:  r.SeasonId,
				CreatedAt: r.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"results": results,
	})
}
