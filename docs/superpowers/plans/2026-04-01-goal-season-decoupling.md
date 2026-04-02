# Goal/Season Decoupling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the hard 1:1 coupling between Goals and Seasons. Goals become top-level team resources; season associations are managed through a new `goal_seasons` join table with two GSIs.

**Architecture:** `Goal.SeasonId` is replaced by `Goal.TeamId`. A new `goal_seasons` DynamoDB table records (goalId, seasonId) pairs. All goal routes move from `/seasons/{seasonId}/goals` to `/teams/{teamId}/goals`. Season-filtered goal listing queries the join table. A one-off migration script backfills existing data.

**Tech Stack:** Go 1.x, AWS DynamoDB (aws-sdk-go-v2), Terraform HCL, testify/assert

---

### Task 1: Update `Goal` model

**Files:**
- Modify: `files/src/models/goals.go`

- [ ] **Step 1: Replace `SeasonId` with `TeamId` in the `Goal` struct**

Replace `files/src/models/goals.go` entirely:

```go
package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type GoalType string
type GoalStatus string

const (
	GoalTypeIndividual GoalType = "individual"
	GoalTypeTeam       GoalType = "team"

	GoalStatusOpen       GoalStatus = "open"
	GoalStatusInProgress GoalStatus = "in_progress"
	GoalStatusCompleted  GoalStatus = "completed"
	GoalStatusArchived   GoalStatus = "archived"
)

type Goal struct {
	Id          string     `dynamodbav:"id" json:"id"`
	TeamId      string     `dynamodbav:"teamId" json:"teamId"`
	OwnerId     string     `dynamodbav:"ownerId" json:"ownerId"`
	GoalType    GoalType   `dynamodbav:"goalType" json:"goalType"`
	Picture     string     `dynamodbav:"picture" json:"picture"`
	Title       string     `dynamodbav:"title" json:"title"`
	Description string     `dynamodbav:"description" json:"description"`
	Status      GoalStatus `dynamodbav:"status" json:"status"`
	CreatedBy   string     `dynamodbav:"createdBy" json:"createdBy"`
	CreatedAt   time.Time  `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time  `dynamodbav:"updatedAt" json:"updatedAt"`
}

func (g *Goal) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(g)
	if err != nil {
		return nil
	}
	return m
}
```

- [ ] **Step 2: Build — expect compile errors in files that reference `goal.SeasonId`**

```bash
cd files/src
go build ./...
```
Expected: compile errors referencing `.SeasonId` in `db/goals.go`, `router/goals/goals_router.go`, `router/comments/comments_router.go`.

Note the files reported — they are fixed in subsequent tasks.

---

### Task 2: Create `GoalSeason` model

**Files:**
- Create: `files/src/models/goal_seasons.go`

- [ ] **Step 1: Create the model file**

```go
package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type GoalSeason struct {
	Id        string    `dynamodbav:"id" json:"id"`
	GoalId    string    `dynamodbav:"goalId" json:"goalId"`
	SeasonId  string    `dynamodbav:"seasonId" json:"seasonId"`
	CreatedAt time.Time `dynamodbav:"createdAt" json:"createdAt"`
}

func (gs *GoalSeason) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(gs)
	if err != nil {
		return nil
	}
	return m
}
```

---

### Task 3: Create `db/goal_seasons.go`

**Files:**
- Create: `files/src/db/goal_seasons.go`

- [ ] **Step 1: Write the goal_seasons DB functions**

```go
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

func TagGoalToSeason(ctx context.Context, goalId, seasonId string) (*models.GoalSeason, error) {
	client = GetClient()
	gs := &models.GoalSeason{
		Id:        goalId + "#" + seasonId,
		GoalId:    goalId,
		SeasonId:  seasonId,
		CreatedAt: time.Now(),
	}
	item, err := attributevalue.MarshalMap(gs)
	if err != nil {
		return nil, fmt.Errorf("TagGoalToSeason: marshal: %w", err)
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &goalSeasonsTableName,
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return nil, fmt.Errorf("TagGoalToSeason: goal %s already tagged to season %s", goalId, seasonId)
		}
		return nil, fmt.Errorf("TagGoalToSeason: put: %w", err)
	}
	return gs, nil
}

func UntagGoalFromSeason(ctx context.Context, goalId, seasonId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &goalSeasonsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: goalId + "#" + seasonId},
		},
	})
	if err != nil {
		return fmt.Errorf("UntagGoalFromSeason: %w", err)
	}
	return nil
}

func ListSeasonsByGoalId(ctx context.Context, goalId string) ([]*models.GoalSeason, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &goalSeasonsTableName,
		IndexName:              aws.String("goalIdIndex"),
		KeyConditionExpression: aws.String("goalId = :gid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gid": &types.AttributeValueMemberS{Value: goalId},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ListSeasonsByGoalId: %w", err)
	}
	var items []*models.GoalSeason
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &items); err != nil {
		return nil, fmt.Errorf("ListSeasonsByGoalId: unmarshal: %w", err)
	}
	return items, nil
}

func ListGoalIdsBySeasonId(ctx context.Context, seasonId string) ([]string, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &goalSeasonsTableName,
		IndexName:              aws.String("seasonIdIndex"),
		KeyConditionExpression: aws.String("seasonId = :sid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": &types.AttributeValueMemberS{Value: seasonId},
		},
		ProjectionExpression: aws.String("goalId"),
	})
	if err != nil {
		return nil, fmt.Errorf("ListGoalIdsBySeasonId: %w", err)
	}
	ids := make([]string, 0, len(result.Items))
	for _, item := range result.Items {
		if v, ok := item["goalId"].(*types.AttributeValueMemberS); ok {
			ids = append(ids, v.Value)
		}
	}
	return ids, nil
}
```

Add `"errors"` to imports.

- [ ] **Step 2: Add `goalSeasonsTableName` to both init files**

In `files/src/db/lambda_init.go`, add inside the `var ()` block:
```go
goalSeasonsTableName = os.Getenv("GOAL_SEASONS_TABLE_NAME")
```

In `files/src/db/local_init.go`, add inside the `var ()` block:
```go
goalSeasonsTableName = "dev-goal-seasons"
```

- [ ] **Step 3: Build — expect remaining compile errors only in router files**

```bash
cd files/src
go build ./...
```
Expected: only errors in `db/goals.go` and router files referencing the old `SeasonId` field.

---

### Task 4: Update `db/goals.go`

**Files:**
- Modify: `files/src/db/goals.go`

- [ ] **Step 1: Update `CreateGoal` — replace `seasonId` param with `teamId`**

Replace the `CreateGoal` function:

```go
func CreateGoal(ctx context.Context, teamId string, ownerId string, goalType models.GoalType, title string, description string) (*models.Goal, error) {
	client = GetClient()
	now := time.Now()
	goal := &models.Goal{
		Id:          models.GenerateID(),
		TeamId:      teamId,
		OwnerId:     ownerId,
		GoalType:    goalType,
		Title:       title,
		Description: description,
		Status:      models.GoalStatusOpen,
		CreatedBy:   ownerId,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &goalsTableName,
		Item:      goal.ToAttributeValues(),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateGoal: %w", err)
	}
	return goal, nil
}
```

Add `"fmt"` to imports if not already present.

- [ ] **Step 2: Update `CountGoalsBySeasonId` — query via `goal_seasons` join table**

`CountGoalsBySeasonId` currently filters goals by `seasonId`. Since goals no longer have a `seasonId` field, rewrite it to first fetch goalIds from the join table:

```go
func CountGoalsBySeasonId(ctx context.Context, seasonId string) (total int, completed int, open int, inProgress int, err error) {
	goalIds, err := ListGoalIdsBySeasonId(ctx, seasonId)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("CountGoalsBySeasonId: list goal IDs: %w", err)
	}
	for _, gid := range goalIds {
		goal, gerr := GetGoalById(ctx, gid)
		if gerr != nil {
			return 0, 0, 0, 0, fmt.Errorf("CountGoalsBySeasonId: get goal %s: %w", gid, gerr)
		}
		if goal == nil || goal.Status == models.GoalStatusArchived {
			continue
		}
		total++
		switch goal.Status {
		case models.GoalStatusCompleted:
			completed++
		case models.GoalStatusOpen:
			open++
		case models.GoalStatusInProgress:
			inProgress++
		}
	}
	return total, completed, open, inProgress, nil
}
```

- [ ] **Step 3: Update `SearchGoalsForTeam` — filter by `goal.TeamId` instead of `goal.SeasonId`**

Replace `SearchGoalsForTeam`:

```go
func SearchGoalsForTeam(ctx context.Context, teamId, query string, limit int) ([]*models.Goal, error) {
	client = GetClient()
	queryLower := strings.ToLower(query)
	results := make([]*models.Goal, 0, limit)
	var lastKey map[string]types.AttributeValue

	for {
		in := &dynamodb.ScanInput{
			TableName:        aws.String(goalsTableName),
			FilterExpression: aws.String("teamId = :tid AND #st <> :archived"),
			ExpressionAttributeNames: map[string]string{
				"#st": "status",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":tid":      &types.AttributeValueMemberS{Value: teamId},
				":archived": &types.AttributeValueMemberS{Value: string(models.GoalStatusArchived)},
			},
		}
		if lastKey != nil {
			in.ExclusiveStartKey = lastKey
		}
		result, scanErr := client.Scan(ctx, in)
		if scanErr != nil {
			return nil, fmt.Errorf("SearchGoalsForTeam: scan: %w", scanErr)
		}
		for _, item := range result.Items {
			if len(results) >= limit {
				break
			}
			titleV, ok := item["title"].(*types.AttributeValueMemberS)
			if !ok {
				continue
			}
			if !strings.Contains(strings.ToLower(titleV.Value), queryLower) {
				continue
			}
			var g models.Goal
			if err := attributevalue.UnmarshalMap(item, &g); err != nil {
				return nil, fmt.Errorf("SearchGoalsForTeam: unmarshal: %w", err)
			}
			results = append(results, &g)
		}
		if len(results) >= limit || result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}
	return results, nil
}
```

- [ ] **Step 4: Remove `GetAllSeasonIdsByTeamId` dependency from `db/goals.go`**

The old `SearchGoalsForTeam` called `GetAllSeasonIdsByTeamId`. With the new implementation filtering directly by `teamId`, this import is no longer used in `goals.go`. Ensure the build passes.

- [ ] **Step 5: Build — expect errors only in router files**

```bash
cd files/src
go build ./...
```
Expected: errors only in `router/goals/goals_router.go` and `router/comments/comments_router.go`.

---

### Task 5: Update `goals_router.go`

**Files:**
- Modify: `files/src/router/goals/goals_router.go`
- Modify: `files/src/router/goals/goals_types.go`

- [ ] **Step 1: Update `goals_types.go` — remove `SeasonId` from request types**

In `files/src/router/goals/goals_types.go`, verify `CreateGoalRequest` does not include `SeasonId` (it already doesn't based on the existing code). No change needed there.

- [ ] **Step 2: Rewrite `CreateGoal` — use `teamId` path param directly**

```go
func CreateGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	var request CreateGoalRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeGoals}, models.PermGoalsWrite) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	callerId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	ownerId := callerId
	if request.OwnerId != nil && utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeGoals}, models.PermGoalsWrite) {
		ownerId = *request.OwnerId
	}
	goal, err := db.CreateGoal(ctx, teamId, ownerId, request.Type, request.Title, request.Description)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{
		"goal": goal,
	})
}
```

- [ ] **Step 3: Rewrite `GetGoal` — use `teamId` path param**

```go
func GetGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	if teamId == "" || goalId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsRead)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"goal": goal,
	})
}
```

- [ ] **Step 4: Rewrite `ListGoals` — use `teamId` path param, add optional `seasonId` query filter**

```go
func ListGoals(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	if teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	filter, err := db.GoalFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	filter.TeamId = teamId

	// Optional season filter: if ?seasonId= is provided, restrict to goals tagged to that season.
	if sid, ok := event.QueryStringParameters["seasonId"]; ok && sid != "" {
		goalIds, err := db.ListGoalIdsBySeasonId(ctx, sid)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		filter.GoalIds = goalIds
		if len(goalIds) == 0 {
			return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
				"items": []interface{}{}, "count": 0, "nextToken": "", "hasMore": false,
			})
		}
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		filter.OwnerId = actorId
		allowed, aerr := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: actorId},
			models.PermGoalsRead)
		if aerr != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	items, count, nextCursor, hasMore, err := db.ListGoals(ctx, filter)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	ownerCache := map[string]*GoalOwner{}
	for _, g := range items {
		ownerCache[g.OwnerId] = nil
	}
	for sub := range ownerCache {
		u, uerr := users.GetUserBySub(ctx, sub)
		if uerr == nil && u != nil {
			ownerCache[sub] = &GoalOwner{
				Id:                u.Id,
				Name:              u.Name,
				PreferredUsername: u.PreferredUsername,
				Picture:           u.Picture,
			}
		}
	}
	goalIds := make([]string, 0, len(items))
	for _, g := range items {
		goalIds = append(goalIds, g.Id)
	}
	progressByGoal, err := db.ListProgressEntriesByGoalIds(ctx, goalIds)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	enriched := make([]GoalWithOwner, 0, len(items))
	for _, g := range items {
		enriched = append(enriched, GoalWithOwner{
			Goal:                 g,
			Owner:                ownerCache[g.OwnerId],
			CompletionPercentage: computeCompletionPercentage(progressByGoal[g.Id]),
		})
	}

	nextToken := ""
	if nextCursor != nil {
		nextToken, err = models.EncodeCursor(nextCursor)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items":     enriched,
		"count":     count,
		"nextToken": nextToken,
		"hasMore":   hasMore,
	})
}
```

- [ ] **Step 5: Update `GoalFilter` in `db/filter.go` (or wherever `GoalFilter` is defined) — add `TeamId` and `GoalIds` fields**

Find the `GoalFilter` struct (search in `files/src/db/`) and add:
```go
TeamId  string
GoalIds []string
```

Update the `BuildExpression` method on `GoalFilter` to include a `teamId = :teamId` condition when `TeamId != ""` and to add a batch-IN condition when `GoalIds` is non-empty (use a DynamoDB `contains` expression or build an OR expression over the id list, up to 100 IDs).

For `GoalIds` filtering, build an expression like:
```go
if len(f.GoalIds) > 0 {
    parts := make([]string, 0, len(f.GoalIds))
    for i, id := range f.GoalIds {
        k := fmt.Sprintf(":gid%d", i)
        vals[k] = &types.AttributeValueMemberS{Value: id}
        parts = append(parts, fmt.Sprintf("id = %s", k))
    }
    exprs = append(exprs, "("+strings.Join(parts, " OR ")+")")
}
```

- [ ] **Step 6: Rewrite `UpdateGoal` and `DeleteGoal` — use `teamId` path param**

```go
func UpdateGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	if teamId == "" || goalId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	var request UpdateGoalRequest
	if err := json.Unmarshal([]byte(event.Body), &request); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsWrite)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	updatedGoal, err := db.UpdateGoal(ctx, goalId, request.OwnerId, request.Title, request.Description, request.Status)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	if request.Status != nil {
		activity.EmitGoalStatusChanged(ctx, teamId, actorId, updatedGoal.Title, *request.Status, goalId)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"goal": updatedGoal,
	})
}

func DeleteGoal(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	if teamId == "" || goalId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsDelete)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	if err := db.DeleteGoal(ctx, goalId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	return utils.SuccessResponse(http.StatusNoContent, utils.MsgSuccess, nil)
}
```

- [ ] **Step 7: Rewrite `UploadGoalFile` — use `teamId` path param**

```go
func UploadGoalFile(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	if teamId == "" || goalId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	filename, ok := event.QueryStringParameters["filename"]
	if !ok || filename == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	contentType, ok := event.QueryStringParameters["contentType"]
	if !ok || contentType == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsWrite)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	presignedUrl, key, err := storage.GeneratePresignedUploadURLForGoalPicture(ctx, goalId, filename, contentType, utils.PresignedURLTimeout)
	if err != nil {
		return nil, err
	}
	publicUrl := storage.GetPublicFileURL(key)

	if err := db.UpdateGoalPicture(ctx, goalId, publicUrl); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"uploadUrl": presignedUrl,
		"key":       key,
		"fileUrl":   publicUrl,
	})
}
```

- [ ] **Step 8: Build — expect errors only in `comments_router.go`**

```bash
cd files/src
go build ./...
```
Expected: only errors in `router/comments/comments_router.go`.

---

### Task 6: Update `comments_router.go` — resolve teamId from `goal.TeamId`

**Files:**
- Modify: `files/src/router/comments/comments_router.go`

- [ ] **Step 1: Update `resolveTeamIdFromTarget` for `CommentTypeGoal`**

In `resolveTeamIdFromTarget`, the `CommentTypeGoal` case currently calls `db.GetTeamIdBySeasonId(goal.SeasonId)`. Replace it:

```go
case models.CommentTypeGoal:
    goal, err := db.GetGoalById(ctx, targetId)
    if err != nil {
        return "", err
    }
    if goal == nil {
        return "", nil
    }
    return goal.TeamId, nil
```

- [ ] **Step 2: Build — clean**

```bash
cd files/src
go build ./...
```
Expected: exits 0, no output.

- [ ] **Step 3: Commit**

```bash
git add files/src/models/goals.go files/src/models/goal_seasons.go \
        files/src/db/goal_seasons.go files/src/db/goals.go \
        files/src/db/lambda_init.go files/src/db/local_init.go \
        files/src/router/goals/goals_router.go \
        files/src/router/comments/comments_router.go
git commit -m "feat: decouple goals from seasons — goals now own teamId directly"
```

---

### Task 7: Add `goal_seasons_router.go` — season tagging endpoints

**Files:**
- Create: `files/src/router/goals/goal_seasons_router.go`

- [ ] **Step 1: Create the router file**

```go
package goals

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

func TagGoalToSeason(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	seasonId := event.PathParameters["seasonId"]
	if teamId == "" || goalId == "" || seasonId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsWrite)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	gs, err := db.TagGoalToSeason(ctx, goalId, seasonId)
	if err != nil {
		return utils.ErrorResponse(http.StatusConflict, "goal already tagged to this season", nil)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{
		"goalSeason": gs,
	})
}

func UntagGoalFromSeason(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	seasonId := event.PathParameters["seasonId"]
	if teamId == "" || goalId == "" || seasonId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsWrite)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	if err := db.UntagGoalFromSeason(ctx, goalId, seasonId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	return utils.SuccessResponse(http.StatusNoContent, utils.MsgSuccess, nil)
}

func ListGoalSeasons(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId := event.PathParameters["teamId"]
	goalId := event.PathParameters["goalId"]
	if teamId == "" || goalId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	goal, err := db.GetGoalById(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	if goal == nil || goal.TeamId != teamId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorNotFound, nil)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		allowed, err := utils.CheckPermission(ctx, actorId, teamId,
			models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
			models.PermGoalsRead)
		if err != nil || !allowed {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}

	items, err := db.ListSeasonsByGoalId(ctx, goalId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, nil)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}
```

- [ ] **Step 2: Build**

```bash
cd files/src
go build ./...
```
Expected: exits 0.

- [ ] **Step 3: Commit**

```bash
git add files/src/router/goals/goal_seasons_router.go
git commit -m "feat: add goal season tagging endpoints (tag/untag/list)"
```

---

### Task 8: Add `goal_seasons` DynamoDB table and new Terraform routes

**Files:**
- Modify: `db.tf`
- Create: `routes_goals.tf`

- [ ] **Step 1: Add `goal_seasons` table to `db.tf`**

Append to `db.tf`:

```hcl
resource "aws_dynamodb_table" "goal_seasons" {
  name         = "${var.prefix}-goal-seasons"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "goalId"
    type = "S"
  }
  attribute {
    name = "seasonId"
    type = "S"
  }

  global_secondary_index {
    name            = "goalIdIndex"
    hash_key        = "goalId"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "seasonIdIndex"
    hash_key        = "seasonId"
    projection_type = "ALL"
  }

  tags = local.tags
}
```

- [ ] **Step 2: Add `GOAL_SEASONS_TABLE_NAME` to `locals.tf`**

In `locals.tf`, add to `lambda_environment_variables`:
```hcl
"GOAL_SEASONS_TABLE_NAME" = aws_dynamodb_table.goal_seasons.name
```

- [ ] **Step 3: Create `routes_goals.tf` with goal and goal-season endpoints**

```hcl
# ─── API Gateway Resources ────────────────────────────────────────────────────

resource "aws_api_gateway_resource" "teams_teamId_goals" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId.id
  path_part   = "goals"
}

resource "aws_api_gateway_resource" "teams_teamId_goals_goalId" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId_goals.id
  path_part   = "{goalId}"
}

resource "aws_api_gateway_resource" "teams_teamId_goals_goalId_seasons" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId_goals_goalId.id
  path_part   = "seasons"
}

resource "aws_api_gateway_resource" "teams_teamId_goals_goalId_seasons_seasonId" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId_goals_goalId_seasons.id
  path_part   = "{seasonId}"
}

resource "aws_api_gateway_resource" "teams_teamId_goals_goalId_picture" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId_goals_goalId.id
  path_part   = "picture"
}

# ─── Goal modules ─────────────────────────────────────────────────────────────

module "create_goal_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["POST"]
  name_overwrite        = "create-goal"
  path_name             = "goals"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "CreateGoal"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:PutItem"]
      Resource = [aws_dynamodb_table.goals.arn]
    }
  ]
}

module "list_goals_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["GET"]
  name_overwrite        = "list-goals"
  path_name             = "goals"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "ListGoals"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:Scan", "dynamodb:Query"]
      Resource = [
        aws_dynamodb_table.goals.arn,
        aws_dynamodb_table.goal_seasons.arn,
        "${aws_dynamodb_table.goal_seasons.arn}/index/seasonIdIndex",
        aws_dynamodb_table.progress.arn,
      ]
    }
  ]
}

module "get_goal_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["GET"]
  name_overwrite        = "get-goal"
  path_name             = "{goalId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "GetGoal"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem"]
      Resource = [aws_dynamodb_table.goals.arn]
    }
  ]
}

module "update_goal_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["PUT"]
  name_overwrite        = "update-goal"
  path_name             = "{goalId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "UpdateGoal"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:UpdateItem"]
      Resource = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.activities.arn]
    }
  ]
}

module "delete_goal_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-goal"
  path_name             = "{goalId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "DeleteGoal"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      Resource = [aws_dynamodb_table.goals.arn]
    }
  ]
}

module "upload_goal_picture_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["POST"]
  name_overwrite        = "upload-goal-picture"
  path_name             = "picture"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId_picture.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "UploadGoalFile"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:UpdateItem", "s3:PutObject"]
      Resource = [aws_dynamodb_table.goals.arn, "${aws_s3_bucket.this.arn}/*"]
    }
  ]
}

# ─── Goal Season tagging modules ─────────────────────────────────────────────

module "tag_goal_season_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["POST"]
  name_overwrite        = "tag-goal-season"
  path_name             = "{seasonId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId_seasons_seasonId.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "TagGoalToSeason"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:PutItem"]
      Resource = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.goal_seasons.arn]
    }
  ]
}

module "untag_goal_season_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["DELETE"]
  name_overwrite        = "untag-goal-season"
  path_name             = "{seasonId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId_seasons_seasonId.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "UntagGoalFromSeason"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      Resource = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.goal_seasons.arn]
    }
  ]
}

module "list_goal_seasons_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["GET"]
  name_overwrite        = "list-goal-seasons"
  path_name             = "seasons"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId_seasons.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "ListGoalSeasons"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:Query"]
      Resource = [
        aws_dynamodb_table.goals.arn,
        aws_dynamodb_table.goal_seasons.arn,
        "${aws_dynamodb_table.goal_seasons.arn}/index/goalIdIndex",
      ]
    }
  ]
}
```

> **Note:** Check `routes_seasons.tf` for any existing goal-related module blocks (e.g. `list_goals_ms`, `create_goal_ms`) that are nested under season paths. Remove them from `routes_seasons.tf` since goals are now managed in `routes_goals.tf`.

- [ ] **Step 4: Verify Terraform plan**

```bash
terraform plan
```
Expected: plan adds `aws_dynamodb_table.goal_seasons` and 9 new Lambda/API Gateway resources. Removes any old season-nested goal modules from `routes_seasons.tf`.

- [ ] **Step 5: Commit**

```bash
git add db.tf locals.tf routes_goals.tf routes_seasons.tf
git commit -m "feat: add goal_seasons DynamoDB table and new /teams/{teamId}/goals routes"
```

---

### Task 9: Data migration script

**Files:**
- Create: `files/src/scripts/migrate_goals/main.go`

- [ ] **Step 1: Create migration script**

```go
//go:build ignore

// migrate_goals migrates all existing Goal records from the old SeasonId model
// to the new TeamId model, creating goal_seasons join records in the process.
//
// Run once per environment after deploying the schema changes:
//
//	go run -tags local files/src/scripts/migrate_goals/main.go
package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/fpgschiba/volleygoals/db"
)

// legacyGoal mirrors the old Goal schema with SeasonId still present.
type legacyGoal struct {
	Id       string `dynamodbav:"id"`
	SeasonId string `dynamodbav:"seasonId"`
	TeamId   string `dynamodbav:"teamId"`
}

func main() {
	ctx := context.Background()
	db.InitClient(nil)
	client := db.GetClient()

	var lastKey map[string]interface{ GetType() string }
	migrated := 0
	skipped := 0

	var startKey map[string]interface{}

	for {
		result, err := client.Scan(ctx, &dynamodb.ScanInput{
			TableName:            aws.String("dev-goals"), // replace with actual table name via env var
			ProjectionExpression: aws.String("id, seasonId, teamId"),
			ExclusiveStartKey:    nil, // set below
		})
		if err != nil {
			log.Fatalf("scan goals: %v", err)
		}
		_ = startKey

		for _, item := range result.Items {
			var g legacyGoal
			if err := attributevalue.UnmarshalMap(item, &g); err != nil {
				log.Printf("unmarshal goal: %v — skipping", err)
				skipped++
				continue
			}
			if g.SeasonId == "" {
				skipped++
				continue
			}
			if g.TeamId != "" {
				// Already migrated
				skipped++
				continue
			}

			teamId, err := db.GetTeamIdBySeasonId(ctx, g.SeasonId)
			if err != nil || teamId == "" {
				log.Printf("could not resolve teamId for goal %s (seasonId=%s): %v — skipping", g.Id, g.SeasonId, err)
				skipped++
				continue
			}

			// 1. Write goal_seasons join record
			if _, err := db.TagGoalToSeason(ctx, g.Id, g.SeasonId); err != nil {
				log.Printf("tag goal %s to season %s: %v — skipping", g.Id, g.SeasonId, err)
				skipped++
				continue
			}

			// 2. Set teamId and clear seasonId on the Goal record
			_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
				TableName: aws.String("dev-goals"),
				Key: map[string]interface{}{
					"id": item["id"],
				},
				UpdateExpression: aws.String("SET teamId = :tid REMOVE seasonId"),
				ExpressionAttributeValues: map[string]interface{}{
					":tid": teamId,
				},
			})
			if err != nil {
				log.Printf("update goal %s: %v — skipping", g.Id, err)
				skipped++
				continue
			}

			migrated++
			log.Printf("migrated goal %s (season %s → team %s)", g.Id, g.SeasonId, teamId)
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		_ = lastKey
	}

	log.Printf("Migration complete: %d migrated, %d skipped", migrated, skipped)
}
```

> **Note:** The migration script uses `aws-sdk-go-v2` directly for the UpdateItem since `db.UpdateGoal` doesn't expose a `teamId` setter. Update the table name constant (`dev-goals`) to read from an environment variable for non-local environments.

- [ ] **Step 2: Build check**

```bash
cd files/src
go build ./...
```
Expected: exits 0 (migration script has `//go:build ignore` tag so it doesn't affect the main build).

- [ ] **Step 3: Commit**

```bash
git add files/src/scripts/migrate_goals/main.go
git commit -m "feat: add goal migration script (SeasonId → TeamId + goal_seasons records)"
```

---

### Task 10: Update API documentation

**Files:**
- Modify: `docs/api/goals.md`
- Modify: `docs/api/seasons.md`
- Modify: `docs/api/comments.md`
- Modify: `docs/api/tenants.md`

- [ ] **Step 1: Update `docs/api/goals.md`**

Replace the entire file with documentation reflecting:
- All goal endpoints are now under `/teams/{teamId}/goals`
- Goal object no longer has `seasonId`; has `teamId` instead
- New season-tagging endpoints: `POST/DELETE /teams/{teamId}/goals/{goalId}/seasons/{seasonId}` and `GET /teams/{teamId}/goals/{goalId}/seasons`
- `GET /teams/{teamId}/goals?seasonId={seasonId}` for season-filtered listing

- [ ] **Step 2: Update `docs/api/seasons.md`**

Add a note that:
- Goals are no longer required to belong to a season
- `GET /seasons/{seasonId}/goals` is maintained as a filtered view via the join table
- Season tagging is managed through `/teams/{teamId}/goals/{goalId}/seasons`

- [ ] **Step 3: Update `docs/api/comments.md`**

Add a note that comments on goals now resolve `teamId` via `goal.teamId` directly (internal implementation detail, no API change).

- [ ] **Step 4: Commit**

```bash
git add docs/api/goals.md docs/api/seasons.md docs/api/comments.md
git commit -m "docs: update API docs for goal/season decoupling"
```
