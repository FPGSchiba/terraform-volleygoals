package router

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/storage"
	"github.com/fpgschiba/volleygoals/users"
	"github.com/fpgschiba/volleygoals/utils"
)

func HealthCheck(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	deps := map[string]string{
		"dynamodb": "ok",
		"s3":       "ok",
		"cognito":  "ok",
	}

	allHealthy := true

	if err := db.CheckHealth(ctx); err != nil {
		deps["dynamodb"] = "error"
		allHealthy = false
	}

	if err := storage.CheckHealth(ctx); err != nil {
		deps["s3"] = "error"
		allHealthy = false
	}

	if err := users.CheckHealth(ctx); err != nil {
		deps["cognito"] = "error"
		allHealthy = false
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	return utils.SuccessResponse(httpStatus, utils.MsgSuccess, map[string]interface{}{
		"status":       status,
		"dependencies": deps,
	})
}
