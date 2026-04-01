# Bugs & Improvements Design
**Date:** 2026-04-01
**Branch:** feature/dynamic-permissions-model

---

## Overview

This document covers the design for 7 bugs and improvements identified during testing. They are grouped into three clusters by scope.

---

## Cluster A — Surgical Fixes

### Issue 3 — Seed Lambda IAM missing `UpdateItem`

**Problem:** `seed_defaults.tf` IAM policy only grants `GetItem`, `PutItem`, `Query` on DynamoDB. `UpsertOwnershipPolicy` calls `UpdateItem` when a policy already exists, causing `terraform apply` to fail on re-seed.

**Fix:** Add `dynamodb:UpdateItem` to `aws_iam_role_policy.seed_defaults_dynamo` for both `ownership_policies` and `role_definitions` table ARNs.

---

### Issue 4 — `CreateTenant` does not create owner `TenantMember`

**Problem:** `db.CreateTenant` persists the tenant record but never creates a `TenantMember` entry for the owner. `IsTenantAdmin` queries `tenant_members`, so the owner cannot administer the tenant.

**Fix:** After the `PutItem` succeeds in `db.CreateTenant`, immediately call `db.AddTenantMember(ctx, tenant.Id, ownerId, TenantMemberRoleAdmin)`. If this call fails, the error is returned to the caller — the tenant record is orphaned but tenant creation is an admin-only operation and recoverable.

---

### Issue 5 — `AddTenantMember` allows duplicate memberships

**Problem:** `AddTenantMember` uses `models.GenerateID()` as PK with no uniqueness check on `(tenantId, userId)`. `GetTenantMemberByUserAndTenant` uses `Limit=1`, yielding nondeterministic results with duplicates.

**Fix:** Replace `models.GenerateID()` with a deterministic composite key `tenantId + "#" + userId`. `PutItem` becomes idempotent — re-adding an active member is a safe overwrite. The GSI query in `GetTenantMemberByUserAndTenant` is unchanged.

---

### Issue 6 — Redundant role definition lookup in `PermissionChecker`

**Problem:** `PermissionChecker.Check` steps 4 and 5 both call `LoadRoleByTenant`. `db.GetRoleDefinitionByTenantAndName` already does an internal tenant→global fallback, making step 5 a redundant second global lookup.

**Fix:** Add `db.GetRoleDefinitionByTenantExact(ctx, tenantId, roleName)` — same query with no global fallback. Add a second field `LoadRoleByTenantExact` to `PermissionChecker`, wired to this new function in `DefaultChecker`. Step 4 of `Check` uses `LoadRoleByTenantExact` (tenant-specific only). Step 5 keeps `LoadRoleByTenant` (with global fallback). This cleanly separates the two lookups with no redundancy.

---

### Issue 7 — Deprecated Lambda runtime `provided.al2`

**Problem:** `seed_defaults.tf` hardcodes `runtime = "provided.al2"`. All other microservice modules inherit the same deprecated default from the module variable. Changing the runtime requires editing every module call.

**Fix:**
1. Add `lambda_runtime = "provided.al2023"` to `locals.tf` alongside `lambda_environment_variables` and `lambda_layer_arns`.
2. Pass `runtime = local.lambda_runtime` in every microservice module call and in `seed_defaults.tf`.
3. Update the Lambda module's `locals.tf` `is_go_build_lambda` condition from `var.runtime == "provided.al2"` to `contains(["provided.al2", "provided.al2023"], var.runtime)` so Go builds still trigger with the new runtime value.

---

## Cluster B — Activity Middleware (Issue 1)

### Problem

Only 2 of 6 `EmitXxx` helpers are wired up (goal status changed, progress report created). Goal create/delete, comment CRUD, season CRUD, progress report update/delete, and team member operations emit no activity.

The existing `Visibility` enum (`all` / `admin_trainer`) is a blunt hardcoded tier that will drift as permissions evolve.

### Activity Model Changes

Add `TargetOwnerId string` to `models.Activity` — the owner of the referenced resource at the time of the event. Remove `Visibility` from active use (keep field in DynamoDB for migration compatibility but stop filtering on it).

### New `EmitXxx` helpers in `activity_router.go`

Add the following alongside existing helpers:
- `EmitGoalCreated`, `EmitGoalDeleted`
- `EmitCommentCreated`, `EmitCommentUpdated`, `EmitCommentDeleted`
- `EmitSeasonCreated`, `EmitSeasonUpdated`, `EmitSeasonDeleted`
- `EmitProgressReportUpdated`, `EmitProgressReportDeleted`

All helpers accept `teamId`, `actorId`, and resource-specific context (title, id, ownerId).

### New `db/instrumented` Package

A new package wraps raw `db.*` write functions. Each wrapper:
1. Calls the underlying `db.*` function.
2. On success, calls the relevant `activity.EmitXxx` helper fire-and-forget.
3. Returns the same signature as the wrapped function — router layer swaps `db.X` for `instrumented.X` with no other changes.

Files:
```
db/instrumented/
  goals.go
  comments.go
  seasons.go
  progress_reports.go
  team_members.go
```

### Permission-Based Activity Filtering

In `GetTeamActivity`, replace the `Visibility` enum filter with resource-level permission checks:

1. **Pre-load once** per invocation: actor's `TeamMember` record, `Team` record (for `tenantId`), all relevant `OwnershipPolicy` records (one per unique `TargetType` in the page), and the actor's `RoleDefinition`.
2. **Per-activity check (in-memory):** Map `TargetType` → its read permission constant (`goals:read`, `comments:read`, `seasons:read`, etc.). Call `CheckPermission` against the pre-loaded data — no additional DB reads.
3. Filter out any activities the caller cannot read.

This yields ~7–8 total DB reads per request regardless of page size, compared to ~80 without pre-loading. The "cache" lives only for the duration of the Lambda invocation — correct and sufficient.

---

## Cluster C — Goal/Season Decoupling (Issue 2)

### Problem

`Goal.SeasonId string` is a required 1:1 field. Goals cannot span seasons, and all goal routes are coupled to a season path.

### Data Model Changes

**`Goal` struct:**
- Remove `SeasonId string`
- Add `TeamId string` (direct team ownership, no longer derived via season)

**New `goal_seasons` DynamoDB table:**
```
id        string  (PK)
goalId    string
seasonId  string
createdAt time.Time
```
GSIs:
- `goalIdIndex` — query all seasons for a goal
- `seasonIdIndex` — query all goal IDs for a season

### API Changes

**Goals primary route moves to `/teams/{teamId}/goals`:**
- `POST   /teams/{teamId}/goals` — create goal (no seasonId)
- `GET    /teams/{teamId}/goals` — list goals for team
- `GET    /teams/{teamId}/goals/{goalId}`
- `PUT    /teams/{teamId}/goals/{goalId}`
- `DELETE /teams/{teamId}/goals/{goalId}`
- `POST   /teams/{teamId}/goals/{goalId}/picture`

**Season-filtered view (kept for backwards compatibility):**
- `GET /seasons/{seasonId}/goals` — queries `goal_seasons` by `seasonIdIndex`, batch-fetches goals

**Season tagging endpoints (new):**
- `POST   /teams/{teamId}/goals/{goalId}/seasons/{seasonId}` — tag goal to season
- `DELETE /teams/{teamId}/goals/{goalId}/seasons/{seasonId}` — untag goal from season
- `GET    /teams/{teamId}/goals/{goalId}/seasons` — list seasons for a goal

### New `db/goal_seasons.go`

Functions:
- `TagGoalToSeason(ctx, goalId, seasonId) error`
- `UntagGoalFromSeason(ctx, goalId, seasonId) error`
- `ListSeasonsByGoalId(ctx, goalId) ([]*models.Season, error)`
- `ListGoalIdsBySeasonId(ctx, seasonId) ([]string, error)`

### Router Changes

- `goals_router.go`: remove all `GetTeamIdBySeasonId` calls; use `teamId` path param directly
- `comments_router.go`: `resolveTeamIdFromTarget` for `CommentTypeGoal` switches from `goal.SeasonId` → `goal.TeamId`
- `progress_reports_router.go`: `GetTeamIdBySeasonId` calls remain (progress reports stay season-scoped)

### Data Migration

For every existing `Goal` with a non-empty `SeasonId`:
1. Write one `goal_seasons` record linking `goalId` → `seasonId`
2. Set `Goal.TeamId` from `GetTeamIdBySeasonId(goal.SeasonId)`
3. Clear `Goal.SeasonId`

Migration runs as a one-off script (similar to `scripts/seed_defaults/main.go`).

---

## API Documentation Updates

The following API doc files in `docs/api/` must be updated to reflect all changes:

| File | Changes |
|------|---------|
| `goals.md` | New routes under `/teams/{teamId}/goals`, new season-tagging endpoints, remove `seasonId` field from goal object |
| `seasons.md` | Document that goals are no longer required to belong to a season; document `GET /seasons/{seasonId}/goals` as a filtered view |
| `tenants.md` | Document that `CreateTenant` now auto-creates the owner as `tenant_admin` |
| `comments.md` | No route changes but note that comments on goals now resolve `teamId` via `goal.teamId` |

---

## Implementation Order

1. Cluster A surgical fixes (low risk, unblock re-seeding immediately)
2. Cluster B activity middleware (self-contained, no schema migration)
3. Cluster C goal/season decoupling (schema migration required, highest risk)
4. API documentation update (after Cluster C is complete)
