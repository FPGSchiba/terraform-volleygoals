# Activity Middleware Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ensure all write operations emit activity log entries, and replace the blunt `Visibility` enum with per-resource permission checks so users only see activities for resources they can read.

**Architecture:** A new `db/instrumented` package wraps each `db.*` write function, calling the underlying function then emitting an activity fire-and-forget. New `EmitXxx` helpers are added to `router/activity`. `GetTeamActivity` pre-loads the caller's permissions once per invocation and filters the activity list in-memory. `models.Activity` gains `TargetOwnerId` to support ownership-based filtering.

**Tech Stack:** Go 1.x, AWS DynamoDB (aws-sdk-go-v2), testify/assert

**Prerequisite:** Complete `2026-04-01-goal-season-decoupling.md` first — this plan wraps `db.CreateGoal(ctx, teamId, ...)` which has the new signature from that plan.

---

### Task 1: Update `models.Activity` — add `TargetOwnerId`

**Files:**
- Modify: `files/src/models/activity.go`

- [ ] **Step 1: Add `TargetOwnerId` field**

```go
package models

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ActivityVisibility string

const (
	ActivityVisibilityAll          ActivityVisibility = "all"
	ActivityVisibilityAdminTrainer ActivityVisibility = "admin_trainer"
)

type Activity struct {
	Id             string             `dynamodbav:"id" json:"id"`
	TeamId         string             `dynamodbav:"teamId" json:"teamId"`
	ActorId        string             `dynamodbav:"actorId" json:"actorId"`
	ActorName      string             `dynamodbav:"actorName" json:"actorName"`
	ActorPicture   string             `dynamodbav:"actorPicture" json:"actorPicture,omitempty"`
	Action         string             `dynamodbav:"action" json:"action"`
	Description    string             `dynamodbav:"description" json:"description"`
	TargetType     string             `dynamodbav:"targetType" json:"targetType,omitempty"`
	TargetId       string             `dynamodbav:"targetId" json:"targetId,omitempty"`
	TargetOwnerId  string             `dynamodbav:"targetOwnerId" json:"targetOwnerId,omitempty"`
	Visibility     ActivityVisibility `dynamodbav:"visibility" json:"visibility"`
	Timestamp      time.Time          `dynamodbav:"timestamp" json:"timestamp"`
}

func (a *Activity) ToAttributeValues() map[string]types.AttributeValue {
	m, err := ToDynamoMap(a)
	if err != nil {
		return nil
	}
	return m
}
```

- [ ] **Step 2: Build**

```bash
cd files/src
go build ./...
```
Expected: exits 0.

---

### Task 2: Add missing `EmitXxx` helpers to `activity_router.go`

**Files:**
- Modify: `files/src/router/activity/activity_router.go`

- [ ] **Step 1: Add `NewActivityWithOwner` helper**

Add after the existing `NewActivity` function:

```go
// NewActivityWithOwner creates an Activity that records the target resource owner,
// enabling permission-based filtering when listing activities.
func NewActivityWithOwner(teamId, actorId, actorName, actorPicture, action, description, targetType, targetId, targetOwnerId string) *models.Activity {
	a := NewActivity(teamId, actorId, actorName, actorPicture, action, description, targetType, targetId, models.ActivityVisibilityAll)
	a.TargetOwnerId = targetOwnerId
	return a
}
```

- [ ] **Step 2: Add goal activity helpers**

```go
func EmitGoalCreated(ctx context.Context, teamId, userId, goalTitle, goalId, ownerId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"goal.created",
		fmt.Sprintf("Goal \"%s\" was created", goalTitle),
		"goal", goalId, ownerId,
	))
}

func EmitGoalDeleted(ctx context.Context, teamId, userId, goalTitle, goalId, ownerId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"goal.deleted",
		fmt.Sprintf("Goal \"%s\" was deleted", goalTitle),
		"goal", goalId, ownerId,
	))
}
```

Update the existing `EmitGoalStatusChanged` to use `NewActivityWithOwner`:

```go
func EmitGoalStatusChanged(ctx context.Context, teamId, userId, goalTitle string, status models.GoalStatus, goalId, ownerId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"goal.status_changed",
		fmt.Sprintf("Goal \"%s\" status changed to %s", goalTitle, string(status)),
		"goal", goalId, ownerId,
	))
}
```

> **Note:** `EmitGoalStatusChanged` gains an `ownerId string` parameter. Update its one call site in `router/goals/goals_router.go` to pass `updatedGoal.OwnerId`.

- [ ] **Step 3: Add comment activity helpers**

```go
func EmitCommentCreated(ctx context.Context, teamId, userId, commentId, targetOwnerId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"comment.created",
		"A comment was posted",
		"comment", commentId, targetOwnerId,
	))
}

func EmitCommentUpdated(ctx context.Context, teamId, userId, commentId, targetOwnerId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"comment.updated",
		"A comment was updated",
		"comment", commentId, targetOwnerId,
	))
}

func EmitCommentDeleted(ctx context.Context, teamId, userId, commentId, targetOwnerId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"comment.deleted",
		"A comment was deleted",
		"comment", commentId, targetOwnerId,
	))
}
```

- [ ] **Step 4: Add season activity helpers**

```go
func EmitSeasonCreated(ctx context.Context, teamId, userId, seasonName, seasonId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"season.created",
		fmt.Sprintf("Season \"%s\" was created", seasonName),
		"season", seasonId, "",
	))
}

func EmitSeasonUpdated(ctx context.Context, teamId, userId, seasonName, seasonId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"season.updated",
		fmt.Sprintf("Season \"%s\" was updated", seasonName),
		"season", seasonId, "",
	))
}

func EmitSeasonDeleted(ctx context.Context, teamId, userId, seasonName, seasonId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"season.deleted",
		fmt.Sprintf("Season \"%s\" was deleted", seasonName),
		"season", seasonId, "",
	))
}
```

- [ ] **Step 5: Add progress report activity helpers**

```go
func EmitProgressReportUpdated(ctx context.Context, teamId, userId, reportId, ownerId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"progress_report.updated",
		"A progress report was updated",
		"progress_report", reportId, ownerId,
	))
}

func EmitProgressReportDeleted(ctx context.Context, teamId, userId, reportId, ownerId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"progress_report.deleted",
		"A progress report was deleted",
		"progress_report", reportId, ownerId,
	))
}
```

Update the existing `EmitProgressReportCreated` to use `NewActivityWithOwner` and accept `ownerId`:

```go
func EmitProgressReportCreated(ctx context.Context, teamId, userId, reportId, ownerId string) {
	u, _ := users.GetUserBySub(ctx, userId)
	actorName, actorPicture := ResolveActorInfo(u)
	db.EmitActivity(ctx, NewActivityWithOwner(
		teamId, userId, actorName, actorPicture,
		"progress_report.created",
		"A progress report was created",
		"progress_report", reportId, ownerId,
	))
}
```

Update its call site in `router/progress-reports/progress_report_router.go` to pass `report.AuthorId` as `ownerId`.

- [ ] **Step 6: Build**

```bash
cd files/src
go build ./...
```
Expected: exits 0.

- [ ] **Step 7: Commit**

```bash
git add files/src/models/activity.go files/src/router/activity/activity_router.go \
        files/src/router/goals/goals_router.go \
        files/src/router/progress-reports/progress_report_router.go
git commit -m "feat: add TargetOwnerId to Activity and new EmitXxx helpers"
```

---

### Task 3: Create `db/instrumented` package

**Files:**
- Create: `files/src/db/instrumented/goals.go`
- Create: `files/src/db/instrumented/comments.go`
- Create: `files/src/db/instrumented/seasons.go`
- Create: `files/src/db/instrumented/progress_reports.go`
- Create: `files/src/db/instrumented/team_members.go`

- [ ] **Step 1: Create `db/instrumented/goals.go`**

```go
package instrumented

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func CreateGoal(ctx context.Context, teamId, actorId string, goalType models.GoalType, title, description string) (*models.Goal, error) {
	goal, err := db.CreateGoal(ctx, teamId, actorId, goalType, title, description)
	if err != nil {
		return nil, err
	}
	activity.EmitGoalCreated(ctx, teamId, actorId, goal.Title, goal.Id, goal.OwnerId)
	return goal, nil
}

func DeleteGoal(ctx context.Context, teamId, actorId, goalId, goalTitle, goalOwnerId string) error {
	if err := db.DeleteGoal(ctx, goalId); err != nil {
		return err
	}
	activity.EmitGoalDeleted(ctx, teamId, actorId, goalTitle, goalId, goalOwnerId)
	return nil
}
```

- [ ] **Step 2: Create `db/instrumented/comments.go`**

```go
package instrumented

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func CreateComment(ctx context.Context, teamId, actorId, commentType, targetId, content string, authorName, authorPicture *string, targetOwnerId string) (*models.Comment, error) {
	comment, err := db.CreateComment(ctx, actorId, commentType, targetId, content, authorName, authorPicture)
	if err != nil {
		return nil, err
	}
	activity.EmitCommentCreated(ctx, teamId, actorId, comment.Id, targetOwnerId)
	return comment, nil
}

func UpdateComment(ctx context.Context, teamId, actorId, commentId, content, targetOwnerId string) (*models.Comment, error) {
	comment, err := db.UpdateComment(ctx, commentId, content)
	if err != nil {
		return nil, err
	}
	activity.EmitCommentUpdated(ctx, teamId, actorId, commentId, targetOwnerId)
	return comment, nil
}

func DeleteComment(ctx context.Context, teamId, actorId, commentId, targetOwnerId string) error {
	if err := db.DeleteComment(ctx, commentId); err != nil {
		return err
	}
	activity.EmitCommentDeleted(ctx, teamId, actorId, commentId, targetOwnerId)
	return nil
}
```

- [ ] **Step 3: Create `db/instrumented/seasons.go`**

```go
package instrumented

import (
	"context"
	"time"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func CreateSeason(ctx context.Context, teamId, actorId, name string, start, end time.Time) (*models.Season, error) {
	season, err := db.CreateSeason(ctx, teamId, name, start, end)
	if err != nil {
		return nil, err
	}
	activity.EmitSeasonCreated(ctx, teamId, actorId, season.Name, season.Id)
	return season, nil
}

func UpdateSeason(ctx context.Context, teamId, actorId, seasonId string, name *string, start, end *time.Time, status *models.SeasonStatus) (*models.Season, error) {
	season, err := db.UpdateSeason(ctx, seasonId, name, start, end, status)
	if err != nil {
		return nil, err
	}
	activity.EmitSeasonUpdated(ctx, teamId, actorId, season.Name, seasonId)
	return season, nil
}

func DeleteSeason(ctx context.Context, teamId, actorId, seasonId, seasonName string) error {
	if err := db.DeleteSeason(ctx, seasonId); err != nil {
		return err
	}
	activity.EmitSeasonDeleted(ctx, teamId, actorId, seasonName, seasonId)
	return nil
}
```

- [ ] **Step 4: Create `db/instrumented/progress_reports.go`**

```go
package instrumented

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func UpdateProgressReport(ctx context.Context, teamId, actorId, reportId string, summary, details, overallDetails *string, entries []db.ProgressEntry, ownerId string) (*models.ProgressReport, error) {
	report, err := db.UpdateProgressReport(ctx, reportId, summary, details, overallDetails, entries)
	if err != nil {
		return nil, err
	}
	activity.EmitProgressReportUpdated(ctx, teamId, actorId, reportId, ownerId)
	return report, nil
}

func DeleteProgressReport(ctx context.Context, teamId, actorId, reportId, ownerId string) error {
	if err := db.DeleteProgressReport(ctx, reportId); err != nil {
		return err
	}
	activity.EmitProgressReportDeleted(ctx, teamId, actorId, reportId, ownerId)
	return nil
}
```

- [ ] **Step 5: Create `db/instrumented/team_members.go`**

```go
package instrumented

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/router/activity"
)

func AddTeamMember(ctx context.Context, teamId, actorId, userId string, role models.TeamMemberRole) (*models.TeamMember, error) {
	member, err := db.AddTeamMember(ctx, teamId, userId, role)
	if err != nil {
		return nil, err
	}
	activity.EmitMemberJoined(ctx, teamId, userId)
	return member, nil
}

func UpdateTeamMemberRole(ctx context.Context, teamId, actorId, memberId string, role models.TeamMemberRole) (*models.TeamMember, error) {
	member, err := db.UpdateTeamMemberRole(ctx, memberId, role)
	if err != nil {
		return nil, err
	}
	activity.EmitMemberRoleChanged(ctx, teamId, actorId, role, memberId)
	return member, nil
}

func RemoveTeamMember(ctx context.Context, teamId, actorId, memberId string) error {
	if err := db.RemoveTeamMember(ctx, memberId); err != nil {
		return err
	}
	activity.EmitMemberRemoved(ctx, teamId, actorId, memberId)
	return nil
}
```

> **Note:** `db.AddTeamMember`, `db.UpdateTeamMemberRole`, and `db.RemoveTeamMember` must exist in the `db` package. Verify their exact signatures in `db/team_members.go` and adjust parameter order if needed.

- [ ] **Step 6: Build**

```bash
cd files/src
go build ./...
```
Expected: exits 0.

- [ ] **Step 7: Commit**

```bash
git add files/src/db/instrumented/
git commit -m "feat: add db/instrumented package that wraps DB writes with activity emission"
```

---

### Task 4: Wire `instrumented.*` into routers

**Files:**
- Modify: `files/src/router/goals/goals_router.go`
- Modify: `files/src/router/comments/comments_router.go`
- Modify: `files/src/router/seasons/seasons_router.go`
- Modify: `files/src/router/progress-reports/progress_report_router.go`
- Modify: `files/src/router/team-members/team_member_router.go`

- [ ] **Step 1: Update `goals_router.go` — use `instrumented.CreateGoal` and `instrumented.DeleteGoal`**

Add import: `"github.com/fpgschiba/volleygoals/db/instrumented"`

In `CreateGoal`, replace:
```go
goal, err := db.CreateGoal(ctx, teamId, ownerId, request.Type, request.Title, request.Description)
```
With:
```go
goal, err := instrumented.CreateGoal(ctx, teamId, actorId, request.Type, request.Title, request.Description)
```

In `DeleteGoal`, before deleting, capture `goal.Title` and `goal.OwnerId`, then replace:
```go
if err := db.DeleteGoal(ctx, goalId); err != nil {
```
With:
```go
if err := instrumented.DeleteGoal(ctx, teamId, actorId, goalId, goal.Title, goal.OwnerId); err != nil {
```

In `UpdateGoal`, update the `EmitGoalStatusChanged` call to pass `updatedGoal.OwnerId`:
```go
if request.Status != nil {
    activity.EmitGoalStatusChanged(ctx, teamId, actorId, updatedGoal.Title, *request.Status, goalId, updatedGoal.OwnerId)
}
```

- [ ] **Step 2: Update `comments_router.go` — use `instrumented` for create/update/delete**

Add import: `"github.com/fpgschiba/volleygoals/db/instrumented"`

The `targetOwnerId` for a comment is the owner of the parent resource (goal owner or report author). Resolve it before the write:

In `CreateComment`, after loading the goal/report to check settings, resolve `targetOwnerId`:
```go
var targetOwnerId string
switch request.CommentType {
case models.CommentTypeGoal:
    g, _ := db.GetGoalById(ctx, request.TargetId)
    if g != nil {
        targetOwnerId = g.OwnerId
    }
case models.CommentTypeProgressReport:
    r, _ := db.GetProgressReportById(ctx, request.TargetId)
    if r != nil {
        targetOwnerId = r.AuthorId
    }
}
```

Then replace `db.CreateComment(...)` with:
```go
comment, err := instrumented.CreateComment(ctx, teamId, authorId, string(request.CommentType), request.TargetId, request.Content, authorName, authorPicture, targetOwnerId)
```

In `UpdateComment`, resolve `targetOwnerId` from the existing comment's target, then replace `db.UpdateComment(...)` with:
```go
updatedComment, err := instrumented.UpdateComment(ctx, teamId, actorId, commentId, request.Content, targetOwnerId)
```

In `DeleteComment`, resolve `targetOwnerId` similarly, then replace `db.DeleteComment(...)` with:
```go
if err := instrumented.DeleteComment(ctx, teamId, actorId, commentId, targetOwnerId); err != nil {
```

- [ ] **Step 3: Update `seasons_router.go` — use `instrumented` for create/update/delete**

Add import: `"github.com/fpgschiba/volleygoals/db/instrumented"`

In `CreateSeason`, replace `db.CreateSeason(...)` with:
```go
season, err := instrumented.CreateSeason(ctx, teamId, actorId, request.Name, request.StartDate, request.EndDate)
```
(where `actorId = utils.GetCognitoUsername(event.RequestContext.Authorizer)`)

In `UpdateSeason`, replace `db.UpdateSeason(...)` with:
```go
season, err := instrumented.UpdateSeason(ctx, teamId, actorId, seasonId, request.Name, request.StartDate, request.EndDate, request.Status)
```

In `DeleteSeason`, fetch the season first (to get `season.Name`), then replace `db.DeleteSeason(...)` with:
```go
if err := instrumented.DeleteSeason(ctx, teamId, actorId, seasonId, season.Name); err != nil {
```

- [ ] **Step 4: Update `progress_report_router.go` — use `instrumented` for update/delete**

Add import: `"github.com/fpgschiba/volleygoals/db/instrumented"`

In `UpdateProgressReport`, replace `db.UpdateProgressReport(...)` with:
```go
updatedReport, err := instrumented.UpdateProgressReport(ctx, teamId, actorId, reportId, request.Summary, request.Details, request.OverallDetails, entries, report.AuthorId)
```
(where `actorId = utils.GetCognitoUsername(event.RequestContext.Authorizer)` and `teamId` is resolved earlier in the function)

In `DeleteProgressReport`, replace `db.DeleteProgressReport(...)` with:
```go
if err := instrumented.DeleteProgressReport(ctx, teamId, actorId, reportId, report.AuthorId); err != nil {
```

- [ ] **Step 5: Update `team_member_router.go` — use `instrumented` for add/update/remove**

Add import: `"github.com/fpgschiba/volleygoals/db/instrumented"`

Replace `db.AddTeamMember(...)`, `db.UpdateTeamMemberRole(...)`, and `db.RemoveTeamMember(...)` with their `instrumented.*` counterparts, passing `teamId` and `actorId` as additional parameters.

- [ ] **Step 6: Build**

```bash
cd files/src
go build ./...
```
Expected: exits 0.

- [ ] **Step 7: Commit**

```bash
git add files/src/router/goals/goals_router.go \
        files/src/router/comments/comments_router.go \
        files/src/router/seasons/seasons_router.go \
        files/src/router/progress-reports/progress_report_router.go \
        files/src/router/team-members/team_member_router.go
git commit -m "feat: wire instrumented DB calls into all write routers"
```

---

### Task 5: Pre-load permission checker for activity listing

**Files:**
- Modify: `files/src/router/activity/activity_router.go`
- Modify: `files/src/utils/check_permission.go`

- [ ] **Step 1: Add `PreloadedChecker` to `utils/check_permission.go`**

Add after `DefaultChecker`:

```go
// PreloadedData holds pre-fetched permission data for a single request.
// All fields are resolved once and reused across multiple Check calls.
type PreloadedData struct {
	Member          *models.TeamMember
	Team            *models.Team
	OwnershipByType map[string]*models.OwnershipPolicy
	RoleExact       *models.RoleDefinition // tenant-specific, may be nil
	RoleGlobal      *models.RoleDefinition // global default
}

// PreloadPermissions fetches all permission data needed to check activities
// for the given actor+team in a single batch of DB reads.
func PreloadPermissions(ctx context.Context, actorId, teamId string) (*PreloadedData, error) {
	member, err := db.GetTeamMemberByUserIDAndTeamID(ctx, actorId, teamId)
	if err != nil {
		return nil, fmt.Errorf("PreloadPermissions: load member: %w", err)
	}
	if member == nil {
		return &PreloadedData{}, nil
	}

	team, err := db.GetTeamById(ctx, teamId)
	if err != nil {
		return nil, fmt.Errorf("PreloadPermissions: load team: %w", err)
	}

	tenantId := ""
	if team != nil && team.TenantId != nil {
		tenantId = *team.TenantId
	}

	resourceTypes := []string{
		models.ResourceTypeGoals,
		models.ResourceTypeComments,
		models.ResourceTypeProgressReports,
		models.ResourceTypeProgress,
		models.ResourceTypeSeasons,
	}
	ownershipByType := make(map[string]*models.OwnershipPolicy, len(resourceTypes))
	for _, rt := range resourceTypes {
		policy, perr := db.GetOwnershipPolicy(ctx, tenantId, rt)
		if perr != nil {
			return nil, fmt.Errorf("PreloadPermissions: load ownership policy for %s: %w", rt, perr)
		}
		ownershipByType[rt] = policy
	}

	var roleExact, roleGlobal *models.RoleDefinition
	roleName := string(member.Role)
	if tenantId != "" {
		roleExact, err = db.GetRoleDefinitionByTenantExact(ctx, tenantId, roleName)
		if err != nil {
			return nil, fmt.Errorf("PreloadPermissions: load tenant role: %w", err)
		}
	}
	roleGlobal, err = db.GetRoleDefinitionByTenantAndName(ctx, "global", roleName)
	if err != nil {
		return nil, fmt.Errorf("PreloadPermissions: load global role: %w", err)
	}

	return &PreloadedData{
		Member:          member,
		Team:            team,
		OwnershipByType: ownershipByType,
		RoleExact:       roleExact,
		RoleGlobal:      roleGlobal,
	}, nil
}

// CanReadActivity returns true if the actor represented by pd can read the given activity.
// All checks are pure in-memory — no DB calls.
func (pd *PreloadedData) CanReadActivity(actorId string, a *models.Activity) bool {
	if pd.Member == nil {
		return false
	}

	readPerm := resourceReadPerm(a.TargetType)
	resource := models.Resource{
		Type:    a.TargetType,
		OwnedBy: a.TargetOwnerId,
	}

	// Ownership check
	if resource.OwnedBy != "" && resource.OwnedBy == actorId {
		if policy, ok := pd.OwnershipByType[resource.Type]; ok && policy != nil {
			if containsString(policy.OwnerPermissions, readPerm) {
				return true
			}
		}
	}

	// Tenant-specific role check
	if pd.RoleExact != nil && containsString(pd.RoleExact.Permissions, readPerm) {
		return true
	}

	// Global role check
	if pd.RoleGlobal != nil && containsString(pd.RoleGlobal.Permissions, readPerm) {
		return true
	}

	return false
}

// resourceReadPerm maps a TargetType string to its canonical read permission constant.
func resourceReadPerm(targetType string) string {
	switch targetType {
	case models.ResourceTypeGoals:
		return models.PermGoalsRead
	case models.ResourceTypeComments:
		return models.PermCommentsRead
	case models.ResourceTypeProgressReports:
		return models.PermProgressReportsRead
	case models.ResourceTypeProgress:
		return models.PermProgressRead
	case models.ResourceTypeSeasons:
		return models.PermSeasonsRead
	default:
		return ""
	}
}
```

Add `"fmt"` to imports in `check_permission.go` if not already present.

- [ ] **Step 2: Update `GetTeamActivity` to use pre-loaded permission filtering**

Replace the visibility-based filter in `GetTeamActivity` in `files/src/router/activity/activity_router.go`:

```go
func GetTeamActivity(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	teamId, ok := event.PathParameters["teamId"]
	if !ok || teamId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}

	if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}

	filter, err := db.ActivityFilterFromQuery(event.QueryStringParameters)
	if err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	filter.TeamId = teamId

	items, count, nextCursor, hasMore, err := db.ListTeamActivities(ctx, filter)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}

	actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		pd, perr := utils.PreloadPermissions(ctx, actorId, teamId)
		if perr != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, perr)
		}
		filtered := items[:0]
		for _, a := range items {
			if pd.CanReadActivity(actorId, a) {
				filtered = append(filtered, a)
			}
		}
		items = filtered
		count = len(items)
	}

	nextToken := ""
	if nextCursor != nil {
		nextToken, err = models.EncodeCursor(nextCursor)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
	}

	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{
		"items":     items,
		"count":     count,
		"nextToken": nextToken,
		"hasMore":   hasMore,
	})
}
```

- [ ] **Step 3: Write tests for `PreloadedData.CanReadActivity`**

Add `files/src/utils/check_permission_preload_test.go`:

```go
package utils_test

import (
	"testing"

	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
	"github.com/stretchr/testify/assert"
)

func preloadedData(memberRole string, ownerPerms, rolePerms []string, resourceType string) *utils.PreloadedData {
	return &utils.PreloadedData{
		Member: &models.TeamMember{Role: models.TeamMemberRole(memberRole)},
		Team:   &models.Team{Id: "team-1"},
		OwnershipByType: map[string]*models.OwnershipPolicy{
			resourceType: {OwnerPermissions: ownerPerms},
		},
		RoleGlobal: &models.RoleDefinition{Permissions: rolePerms},
	}
}

func TestCanReadActivity_OwnerCanReadOwnGoalActivity(t *testing.T) {
	pd := preloadedData("member", []string{models.PermGoalsRead}, nil, models.ResourceTypeGoals)
	a := &models.Activity{TargetType: models.ResourceTypeGoals, TargetOwnerId: "user-1"}
	assert.True(t, pd.CanReadActivity("user-1", a))
}

func TestCanReadActivity_NonOwnerMemberCannotReadGoalActivity(t *testing.T) {
	pd := preloadedData("member", []string{models.PermGoalsRead}, nil, models.ResourceTypeGoals)
	a := &models.Activity{TargetType: models.ResourceTypeGoals, TargetOwnerId: "user-2"}
	assert.False(t, pd.CanReadActivity("user-1", a))
}

func TestCanReadActivity_TrainerCanReadAnyGoalActivity(t *testing.T) {
	pd := preloadedData("trainer", nil, []string{models.PermGoalsRead}, models.ResourceTypeGoals)
	a := &models.Activity{TargetType: models.ResourceTypeGoals, TargetOwnerId: "user-2"}
	assert.True(t, pd.CanReadActivity("trainer-1", a))
}

func TestCanReadActivity_NilMemberDenies(t *testing.T) {
	pd := &utils.PreloadedData{}
	a := &models.Activity{TargetType: models.ResourceTypeGoals, TargetOwnerId: "user-1"}
	assert.False(t, pd.CanReadActivity("user-1", a))
}

func TestCanReadActivity_SeasonActivityVisibleToAllMembers(t *testing.T) {
	pd := preloadedData("member", nil, []string{models.PermSeasonsRead}, models.ResourceTypeSeasons)
	a := &models.Activity{TargetType: models.ResourceTypeSeasons, TargetOwnerId: ""}
	assert.True(t, pd.CanReadActivity("member-1", a))
}
```

- [ ] **Step 4: Run all tests**

```bash
cd files/src
go test ./utils/... -v
```
Expected: all tests pass including the 5 new preload tests.

- [ ] **Step 5: Build final check**

```bash
go build ./...
```
Expected: exits 0.

- [ ] **Step 6: Commit**

```bash
git add files/src/utils/check_permission.go \
        files/src/utils/check_permission_preload_test.go \
        files/src/router/activity/activity_router.go
git commit -m "feat: replace visibility enum with permission-based activity filtering using pre-loaded data"
```

---

### Task 6: Update API documentation

**Files:**
- Modify: `docs/api/tenants.md`

- [ ] **Step 1: Update `docs/api/tenants.md`**

Document that `POST /tenants` now auto-creates the requesting user as a `tenant_admin` member. No separate member creation step is required.

- [ ] **Step 2: Commit**

```bash
git add docs/api/tenants.md
git commit -m "docs: update tenants API docs to reflect auto admin member creation"
```
