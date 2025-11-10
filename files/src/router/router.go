package router

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/router/invites"
	teammembers "github.com/fpgschiba/volleygoals/router/team-members"
	teamsettings "github.com/fpgschiba/volleygoals/router/team-settings"
	"github.com/fpgschiba/volleygoals/router/teams"
	"github.com/fpgschiba/volleygoals/utils"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

// SelectedHandler is set via -ldflags: -X github.com/fpgschiba/volleygoals/router.SelectedHandler=GetTeam
// or left empty to use Router mode. It can also be overridden via HANDLER env var.
var SelectedHandler string
var tracer = otel.Tracer("volleygoals/router")

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC RECOVERED: %v\n", r)
			fmt.Printf("STACK TRACE:\n%s\n", debug.Stack())
		}
	}()

	// Extract X-Ray trace context
	if traceHeader := os.Getenv("_X_AMZN_TRACE_ID"); traceHeader != "" {
		carrier := propagation.MapCarrier{
			"X-Amzn-Trace-Id": traceHeader,
		}
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	}

	h := SelectedHandler
	if env := os.Getenv("HANDLER"); env != "" {
		h = env
	}

	// Use a more descriptive span name
	spanName := fmt.Sprintf("%s %s", event.HTTPMethod, event.Path)
	ctx, span := tracer.Start(ctx, spanName)
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", event.HTTPMethod),
		attribute.String("http.path", event.Path),
		attribute.String("http.route", event.Resource),
		attribute.String("handler", h),
	)

	var response *events.APIGatewayProxyResponse
	var err error

	switch h {
	// Teams handlers
	case "GetTeam":
		response, err = teams.GetTeam(ctx, event)
	case "ListTeams":
		response, err = teams.ListTeams(ctx, event)
	case "CreateTeam":
		response, err = teams.CreateTeam(ctx, event)
	case "UpdateTeam":
		response, err = teams.UpdateTeam(ctx, event)
	case "DeleteTeam":
		response, err = teams.DeleteTeam(ctx, event)

	// Team settings handlers
	case "GetTeamSettings":
		response, err = teamsettings.GetTeamSettings(ctx, event)
	case "UpdateTeamSettings":
		response, err = teamsettings.UpdateTeamSettings(ctx, event)

	// Self handlers
	case "GetSelf":
		response, err = GetSelf(ctx, event)
	case "UpdateSelf":
		response, err = UpdateSelf(ctx, event)

	// Team members handlers
	case "ListTeamMembers":
		response, err = teammembers.ListTeamMembers(ctx, event)
	case "AddTeamMember":
		response, err = teammembers.AddTeamMember(ctx, event)
	case "UpdateTeamMember":
		response, err = teammembers.UpdateTeamMember(ctx, event)
	case "RemoveTeamMember":
		response, err = teammembers.RemoveTeamMember(ctx, event)
	case "LeaveTeam":
		response, err = teammembers.LeaveTeam(ctx, event)

	// Invites handlers
	case "CreateInvite":
		response, err = invites.CreateInvite(ctx, event)
	case "AcceptInvite":
		response, err = invites.AcceptInvite(ctx, event)
	case "ListInvites":
		response, err = invites.ListInvites(ctx, event)
	case "RevokeInvite":
		response, err = invites.RevokeInvite(ctx, event)
	case "ResendInvite":
		response, err = invites.ResendInvite(ctx, event)

	// Unknown handler
	default:
		log.Printf("unknown handler selected: %q", h)
		response, err = utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
	}

	if response != nil {
		span.SetAttributes(attribute.Int("http.status_code", response.StatusCode))
	}

	return response, err
}
