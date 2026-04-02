package seasons

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/db/instrumented"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

func CreateSeason(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var body CreateSeasonRequest
	err := json.Unmarshal([]byte(event.Body), &body)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, body.TeamId, models.Resource{Type: models.ResourceTypeSeasons}, models.PermSeasonsWrite) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	season, err := instrumented.CreateSeason(ctx, body.TeamId, actorId, body.Name, body.StartDate, body.EndDate)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{
		"season": season,
	})
}

func GetSeason(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	authorized, exists, err := isAuthorizedForSeason(ctx, event.RequestContext.Authorizer, seasonId, true)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !exists {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorSeasonNotFound, nil)
	}
	if !authorized {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	season, err := db.GetSeasonById(ctx, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"season": season,
	})
}

func ListSeasons(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	filter, err := db.SeasonFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}

	if filter.TeamId != "" {
		// Authorization: all team users can list seasons
		if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, filter.TeamId, models.Resource{Type: models.ResourceTypeSeasons}, models.PermSeasonsRead) {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	} else {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	items, count, nextCursor, hasMore, err := db.ListSeasons(ctx, filter)
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
		Items:     items,
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

func UpdateSeason(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	authorized, exists, err := isAuthorizedForSeason(ctx, event.RequestContext.Authorizer, seasonId, false)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !exists {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorSeasonNotFound, nil)
	}
	if !authorized {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var body UpdateSeasonRequest
	err = json.Unmarshal([]byte(event.Body), &body)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	existingSeason, err := db.GetSeasonById(ctx, seasonId)
	if err != nil || existingSeason == nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	season, err := instrumented.UpdateSeason(ctx, existingSeason.TeamId, actorId, seasonId, body.Name, body.StartDate, body.EndDate, body.Status)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"season": season,
	})
}

func DeleteSeason(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	seasonId := event.PathParameters["seasonId"]
	authorized, exists, err := isAuthorizedForSeason(ctx, event.RequestContext.Authorizer, seasonId, false)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !exists {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorSeasonNotFound, nil)
	}
	if !authorized {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	deleteSeason, err := db.GetSeasonById(ctx, seasonId)
	if err != nil || deleteSeason == nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	deleteActorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	err = instrumented.DeleteSeason(ctx, deleteSeason.TeamId, deleteActorId, seasonId, deleteSeason.Name)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}

func GetSeasonStats(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
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

	if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeSeasons}, models.PermSeasonsRead) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	goalCount, completedGoalCount, openGoalCount, inProgressGoalCount, err := db.CountGoalsBySeasonId(ctx, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	reportCount, err := db.CountProgressReportsBySeasonId(ctx, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	members, err := db.GetMembershipsByTeamID(ctx, teamId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"stats": map[string]interface{}{
			"goalCount":           goalCount,
			"completedGoalCount":  completedGoalCount,
			"openGoalCount":       openGoalCount,
			"inProgressGoalCount": inProgressGoalCount,
			"reportCount":         reportCount,
			"memberCount":         len(members),
		},
	})
}

func isAuthorizedForSeason(ctx context.Context, authorizer map[string]interface{}, seasonId string, teamUser bool) (bool, bool, error) {
	season, err := db.GetSeasonById(ctx, seasonId)
	if err != nil {
		return false, false, err
	}
	if season == nil {
		return false, false, nil
	}
	// Authorization: only admins or team admins/trainers can access season details
	if teamUser {
		if !utils.HasTeamPermission(ctx, authorizer, season.TeamId, models.Resource{Type: models.ResourceTypeSeasons}, models.PermSeasonsRead) {
			return false, true, nil
		}
	} else {
		if !utils.HasTeamPermission(ctx, authorizer, season.TeamId, models.Resource{Type: models.ResourceTypeSeasons}, models.PermSeasonsWrite) {
			return false, true, nil
		}
	}
	return true, true, nil
}
