package utils

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {
	InitLogger()
}

// InitLogger configures the global logrus logger (JSON output, stdout, level from LOG_LEVEL).
func InitLogger() {
	// Use JSON formatter for all logs so they're consistent for CloudWatch analysis
	log.SetFormatter(&log.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	log.SetOutput(os.Stdout)

	// Allow overriding log level with environment variable LOG_LEVEL (e.g., debug, info, warn, error)
	if lvlStr := os.Getenv("LOG_LEVEL"); lvlStr != "" {
		if lvl, err := log.ParseLevel(strings.ToLower(lvlStr)); err == nil {
			log.SetLevel(lvl)
		} else {
			log.SetLevel(log.InfoLevel)
		}
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func getDurationInMilliseconds(start time.Time) float64 {
	end := time.Now()
	duration := end.Sub(start)
	milliseconds := float64(duration) / float64(time.Millisecond)
	rounded := float64(int(milliseconds*100+.5)) / 100
	return rounded
}

// headerFirst returns the first non-empty header value from Gin for the given names.
func headerFirst(c *gin.Context, names ...string) string {
	for _, n := range names {
		if v := c.GetHeader(n); v != "" {
			return v
		}
	}
	return ""
}

// collectGinErrors joins Gin context errors into a single string (or empty).
func collectGinErrors(c *gin.Context) string {
	if len(c.Errors) == 0 {
		return ""
	}
	errMsgs := make([]string, 0, len(c.Errors))
	for _, e := range c.Errors {
		if e.Err != nil {
			errMsgs = append(errMsgs, e.Error())
		}
	}
	return strings.Join(errMsgs, "; ")
}

// buildFieldsFromGin constructs a log.Fields map from the Gin context and duration (ms float64).
func buildFieldsFromGin(c *gin.Context, duration float64) log.Fields {
	return log.Fields{
		"duration":   duration,
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"status":     c.Writer.Status(),
		"referrer":   headerFirst(c, "Referer", "referer"),
		"request_id": headerFirst(c, "X-Request-Id", "x-request-id"),
	}
}

// classLevel logs an entry depending on status and errors
func classLevel(entry *log.Entry, status int, errStr string) {
	if errStr != "" {
		entry = entry.WithField("errors", errStr)
		entry.Error("handler error")
		return
	}
	if status >= 500 {
		entry.Error("server error")
	} else if status >= 400 {
		entry.Warn("client error")
	} else {
		entry.Info("request handled")
	}
}

// JSONLogLambdaWrapper wraps a Lambda-style handler (ctx, event) -> (response, error)
// and logs duration, method, path, status, referrer and any error using logrus.
func JSONLogLambdaWrapper(handler func(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error)) func(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	return func(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
		start := time.Now()
		resp, err := handler(ctx, event)
		duration := getDurationInMilliseconds(start)

		// safe fetch for referer header (case-insensitive)
		ref := ""
		if v, ok := event.Headers["Referer"]; ok {
			ref = v
		} else if v, ok := event.Headers["referer"]; ok {
			ref = v
		}

		status := 0
		if resp != nil {
			status = resp.StatusCode
		}

		requestID := ""
		if event.RequestContext.RequestID != "" {
			requestID = event.RequestContext.RequestID
		} else if rid, ok := event.Headers["X-Request-Id"]; ok {
			requestID = rid
		}

		fields := log.Fields{
			"duration":   duration,
			"method":     event.HTTPMethod,
			"path":       event.Path,
			"status":     status,
			"referrer":   ref,
			"request_id": requestID,
		}

		entry := log.WithFields(fields)

		if err != nil {
			entry = entry.WithError(err)
			entry.Error("handler error")
		} else if status >= 500 {
			entry.Error("server error")
		} else if status >= 400 {
			entry.Warn("client error")
		} else {
			entry.Info("request handled")
		}

		return resp, err
	}
}

// JSONLogMiddleware returns a gin.HandlerFunc that logs the same fields as
// the Lambda wrapper so logs are consistent between local and Lambda runs.
func JSONLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		// after handler
		duration := getDurationInMilliseconds(start)

		fields := buildFieldsFromGin(c, duration)
		entry := log.WithFields(fields)
		status := fields["status"].(int)
		errStr := collectGinErrors(c)

		classLevel(entry, status, errStr)
	}
}
