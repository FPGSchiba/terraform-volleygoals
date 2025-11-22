//go:build local

package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/mail"
	"github.com/fpgschiba/volleygoals/router"
	"github.com/fpgschiba/volleygoals/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// AuthMiddleware validates Authorization header and stores authorizer claims
// into gin context under ctxAuthorizerKey. On failure it aborts with 401.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read header
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}
		claims, err := utils.ValidateToken(auth)
		if err != nil {
			log.WithError(err).Warn("token validation failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// Build authorizer object similar to API Gateway shape
		authorizer := map[string]interface{}{
			"claims": claims,
		}
		// Save into gin context
		c.Set(utils.CtxAuthorizerKey, authorizer)
		c.Next()
	}
}

// readBody reads the raw body and returns it as string (or empty string when none).
func readBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}
	b, _ := io.ReadAll(c.Request.Body)
	if len(b) == 0 {
		return ""
	}
	return string(b)
}

// copyHeaders copies the first header values from the Gin request into the event.
func copyHeaders(event *events.APIGatewayProxyRequest, c *gin.Context) {
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			event.Headers[k] = v[0]
		}
	}
}

// copyQueryParams copies the first query param values into the event.
func copyQueryParams(event *events.APIGatewayProxyRequest, c *gin.Context) {
	q := c.Request.URL.Query()
	for k, v := range q {
		if len(v) > 0 {
			event.QueryStringParameters[k] = v[0]
		}
	}
}

// copyPathParams copies Gin path parameters into the event.PathParameters.
func copyPathParams(event *events.APIGatewayProxyRequest, c *gin.Context) {
	for _, p := range c.Params {
		event.PathParameters[p.Key] = p.Value
	}
}

// injectAuthorizer copies the authorizer map from Gin context into
// event.RequestContext.Authorizer (if present).
func injectAuthorizer(event *events.APIGatewayProxyRequest, c *gin.Context) {
	if authV, ok := c.Get(utils.CtxAuthorizerKey); ok {
		if aMap, ok2 := authV.(map[string]interface{}); ok2 {
			if event.RequestContext.Authorizer == nil {
				event.RequestContext.Authorizer = map[string]interface{}{}
			}
			for k, v := range aMap {
				event.RequestContext.Authorizer[k] = v
			}
		}
	}
}

// buildEventFromContext builds a minimal APIGatewayProxyRequest from Gin
// context and the provided body string.
func buildEventFromContext(c *gin.Context, bodyStr string) events.APIGatewayProxyRequest {
	event := events.APIGatewayProxyRequest{
		HTTPMethod:            c.Request.Method,
		Path:                  c.Request.URL.Path,
		Resource:              c.FullPath(),
		Body:                  bodyStr,
		Headers:               map[string]string{},
		QueryStringParameters: map[string]string{},
		PathParameters:        map[string]string{},
	}

	copyHeaders(&event, c)
	copyQueryParams(&event, c)
	copyPathParams(&event, c)
	injectAuthorizer(&event, c)

	return event
}

// Adapter is a generic factory that returns a gin.HandlerFunc which converts
// the incoming Gin request into an APIGatewayProxyRequest, calls the
// lambda-style handler via router.HandleRequestWithHandler, and writes the
// APIGatewayProxyResponse back to the client.
func Adapter(handlerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authentication placeholder: implement auth checks here if required.
		// For now this middleware path allows all requests through.

		ctx := context.Background()

		// Ensure DB client is initialized when running locally
		db.InitClient(nil)

		bodyStr := readBody(c)
		event := buildEventFromContext(c, bodyStr)

		resp, err := router.HandleRequestWithHandler(ctx, event, handlerName)
		if err != nil {
			log.WithFields(log.Fields{"handler": handlerName}).WithError(err).Warn("handler returned error")
		}

		if resp == nil {
			c.String(http.StatusInternalServerError, "handler returned error: %v", err)
			return
		}

		for k, v := range resp.Headers {
			c.Header(k, v)
		}

		contentType := resp.Headers["Content-Type"]
		if contentType == "" {
			contentType = "application/json"
		}

		c.Data(resp.StatusCode, contentType, []byte(resp.Body))
	}
}

// CORSMiddleware returns a permissive CORS middleware. Replace with your
// project's CORS implementation if you have special needs.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// GetRouter constructs the Gin engine with groups/routes as requested. Each
// route uses Adapter[...] to forward requests to the lambda-style handlers.
func GetRouter() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(utils.JSONLogMiddleware())
	engine.Use(CORSMiddleware())

	// Public auth endpoints (not under apiGroup to mirror your example)
	// Invite completion endpoint (no auth)
	engine.POST("/api/v1/invites/complete", Adapter("CompleteInvite"))

	apiGroup := engine.Group("/api/v1")
	{
		// Apply auth for subsequent routes
		apiGroup.Use(AuthMiddleware())

		selfGroup := apiGroup.Group("/self") // User only
		{
			selfGroup.GET("/", Adapter("GetSelf"))
			selfGroup.PATCH("/", Adapter("UpdateSelf"))
		}
		teamsGroup := apiGroup.Group("/teams")
		{
			teamsGroup.POST("/", Adapter("CreateTeam")) // Admin only
			teamsGroup.GET("/", Adapter("ListTeams"))   // Admin only
			teamGroup := teamsGroup.Group(":teamId")    // Admin or User for specific Team
			{
				teamGroup.PATCH("/settings", Adapter("UpdateTeamSettings")) // Admin or User with Role Trainer on Team
				teamGroup.DELETE("/", Adapter("DeleteTeam"))                // Admin only
				teamGroup.GET("/", Adapter("GetTeam"))                      // Admin or User for Team
				teamGroup.PATCH("/", Adapter("UpdateTeam"))                 // Admin or User with Role Trainer on Team
				membersGroup := teamGroup.Group("/members")
				{
					membersGroup.POST("/", Adapter("AddTeamMember"))              // Admin only
					membersGroup.GET("/", Adapter("ListTeamMembers"))             // Admin or User for Team
					membersGroup.DELETE("/", Adapter("LeaveTeam"))                // User only and Trainers only if another Trainer exists
					membersGroup.DELETE(":memberId", Adapter("RemoveTeamMember")) // Admin and User with Role Trainer on Team
					membersGroup.PATCH(":memberId", Adapter("UpdateTeamMember"))  // Admin and User with Role Trainer on Team
				}
			}
		}
		invitesGroup := apiGroup.Group("/invites") // Admin or User with Role Trainer on Team
		{
			invitesGroup.POST("/", Adapter("CreateInvite"))
			invitesGroup.GET("/", Adapter("ListInvites"))
			invitesGroup.DELETE(":inviteId", Adapter("RevokeInvite"))
			invitesGroup.PATCH(":inviteId", Adapter("ResendInvite"))
		}
		usersGroup := apiGroup.Group("/users") // Admin only
		{
			usersGroup.GET("/", Adapter("ListUsers"))
			usersGroup.GET("/:userSub", Adapter("GetUser"))
			usersGroup.DELETE(":userSub", Adapter("RemoveUser"))
		}
	}

	// Configurations for the gin router
	engine.RemoveExtraSlash = true
	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = false

	return engine
}

func main() {
	// Ensure logger is configured early
	utils.InitLogger()

	r := GetRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	db.InitClient(nil)
	mail.InitClient(nil)
	log.Infof("starting volleygoals local server on :%s (use /api/v1/... endpoints)", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
