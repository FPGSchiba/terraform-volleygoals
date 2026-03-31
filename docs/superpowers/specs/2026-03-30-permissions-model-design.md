# Dynamic Permissions Model Design

**Date:** 2026-03-30
**Status:** Approved
**Scope:** Team-level permissions with lightweight tenant (club) introduction

---

## Overview

Replace the current hard-coded role checks (`IsTeamAdmin()`, `IsTeamTrainer()`, etc.) with a dynamic RBAC + ownership policy model. Roles carry configurable `resource:action` permission sets stored in DynamoDB. Ownership grants automatic access to a user's own resources. A lightweight Tenant (club) entity is introduced now to avoid rework when multi-tenant onboarding is built.

---

## Background: Current Model

Two permission layers exist today:

1. **Global (Cognito groups):** `ADMINS` and `USERS` — unchanged by this design.
2. **Team-level (DynamoDB):** `admin`, `trainer`, `member` roles on the `team_members` table. Checked via hard-coded functions in `utils/permissions.go` scattered across router handlers.

Limitations:
- Roles and their allowed actions are hard-coded in Go.
- No way to customise permissions per club.
- No ownership-aware access (e.g. a member can't access their own goals without a role check).
- No tenant model to support club onboarding.

---

## Architecture

### Permission Evaluation Chain

A single function replaces all existing permission helper calls:

```go
CheckPermission(ctx context.Context, actorId, teamId string, resource Resource, action string) bool

type Resource struct {
    Type          string // "goals", "comments", "teams", etc.
    OwnedBy       string // direct owner (creator) of the resource
    ParentOwnedBy string // owner of parent resource (e.g. goal owner for a comment)
}
```

Evaluation order (first match wins):

1. Load the actor's `TeamMember` record → get role name. Load the `Team` record to resolve `tenantId` (used in steps 2–5).
2. **Ownership check (direct):** if `resource.OwnedBy == actorId`, load `OwnershipPolicy` for `(tenantId, resourceType)`. If `action ∈ ownerPermissions` → **ALLOW**.
3. **Ownership check (parent):** if `resource.ParentOwnedBy == actorId`, load `OwnershipPolicy` for `(tenantId, resourceType)`. If `action ∈ parentOwnerPermissions` → **ALLOW**.
4. **Role check (tenant-specific):** load `RoleDefinition` for `(tenantId, roleName)`. If `action ∈ permissions` → **ALLOW**.
5. **Role check (global default):** load `RoleDefinition` for `(tenantId=null, roleName)`. If `action ∈ permissions` → **ALLOW**.
6. → **DENY**

### Multiple Roles

Users hold a single role name. Tenant admins can define composite roles (e.g. `trainer-admin`) that contain the union of permission sets from both roles. No multi-role assignment is needed.

---

## Data Model

### New DynamoDB Tables

#### `{prefix}-tenants`

| Field | Type | Notes |
|---|---|---|
| `id` | string | PK |
| `name` | string | Club/tenant name |
| `ownerId` | string | Cognito userId of tenant admin |
| `createdAt` | time | |
| `updatedAt` | time | |

#### `{prefix}-tenant-members`

| Field | Type | Notes |
|---|---|---|
| `id` | string | PK |
| `tenantId` | string | GSI: `tenantIdIndex` |
| `userId` | string | GSI: `userIdIndex` |
| `role` | string | `tenant_admin` \| `tenant_member` |
| `status` | string | `active` \| `removed` |
| `createdAt` | time | |
| `updatedAt` | time | |

Composite GSI: `tenantUserIndex` (`tenantId` + `userId`).

#### `{prefix}-role-definitions`

| Field | Type | Notes |
|---|---|---|
| `id` | string | PK |
| `tenantId` | string | GSI: `tenantIdIndex`. `null` = global default |
| `name` | string | e.g. `trainer`, `admin`, `trainer-admin` |
| `permissions` | []string | e.g. `["goals:read", "goals:write"]` |
| `isDefault` | bool | Whether this is a seed/default role |
| `createdAt` | time | |
| `updatedAt` | time | |

#### `{prefix}-ownership-policies`

| Field | Type | Notes |
|---|---|---|
| `id` | string | PK |
| `tenantId` | string | GSI: `tenantIdIndex`. `null` = global default |
| `resourceType` | string | e.g. `goals`, `comments` |
| `ownerPermissions` | []string | Permissions granted to direct owner |
| `parentOwnerPermissions` | []string | Permissions granted to parent resource owner |
| `createdAt` | time | |
| `updatedAt` | time | |

Composite GSI: `tenantResourceTypeIndex` (`tenantId` + `resourceType`).

### Changes to Existing Tables

#### `{prefix}-teams`
- Add `tenantId` attribute (nullable for backward compatibility).
- Add GSI: `tenantIdIndex` on `tenantId`.

#### `{prefix}-team-members`
- `role` field remains a string but now references a `RoleDefinition.name` within the team's tenant scope rather than the fixed Go enum.
- No structural change to the table.

---

## Default Permission Seed

Global default `RoleDefinition` records (`tenantId = null`) seeded at deploy time:

| Permission | admin | trainer | member |
|---|:---:|:---:|:---:|
| `teams:read` | ✅ | ✅ | ✅ |
| `teams:write` | ✅ | ❌ | ❌ |
| `teams:delete` | ✅ | ❌ | ❌ |
| `team_settings:read` | ✅ | ✅ | ❌ |
| `team_settings:write` | ✅ | ❌ | ❌ |
| `members:read` | ✅ | ✅ | ✅ |
| `members:write` | ✅ | ❌ | ❌ |
| `members:delete` | ✅ | ❌ | ❌ |
| `invites:read` | ✅ | ❌ | ❌ |
| `invites:write` | ✅ | ❌ | ❌ |
| `invites:delete` | ✅ | ❌ | ❌ |
| `seasons:read` | ✅ | ✅ | ✅ |
| `seasons:write` | ❌ | ✅ | ❌ |
| `seasons:delete` | ❌ | ✅ | ❌ |
| `goals:read` *(all members')* | ❌ | ✅ | ❌ |
| `goals:write` | ❌ | ✅ | ❌ |
| `goals:delete` | ❌ | ✅ | ❌ |
| `progress_reports:read` *(all members')* | ❌ | ✅ | ❌ |
| `progress_reports:write` | ❌ | ✅ | ❌ |
| `progress_reports:delete` | ❌ | ✅ | ❌ |
| `progress:read` | ❌ | ✅ | ❌ |
| `progress:write` | ❌ | ✅ | ❌ |
| `comments:read` | ❌ | ✅ | ❌ |
| `comments:write` | ❌ | ✅ | ❌ |
| `comments:delete` | ❌ | ✅ | ❌ |
| `activities:read` | ✅ | ✅ | ❌ |

Global default `OwnershipPolicy` records (`tenantId = null`) seeded at deploy time:

| Resource | `ownerPermissions` | `parentOwnerPermissions` |
|---|---|---|
| `goals` | `goals:read`, `goals:write`, `goals:delete`, `comments:read`, `comments:write` | — |
| `progress_reports` | `progress_reports:read`, `progress_reports:write`, `progress_reports:delete`, `comments:read`, `comments:write` | — |
| `progress` | `progress:read`, `progress:write` | — |
| `comments` | `comments:read`, `comments:write`, `comments:delete` | `comments:read`, `comments:write` |

> Members access their own goals, progress reports, and the comments on them via ownership — not via role. Trainers access all members' athletic data via role. Admins manage roster, settings, and invites but do not access other members' goals or reports.

---

## New API Endpoints

### Tenant Management

```
POST   /v1/tenants                                         CreateTenant          (global ADMINS only)
GET    /v1/tenants/{tenantId}                              GetTenant             (global ADMINS or tenant member)
PATCH  /v1/tenants/{tenantId}                              UpdateTenant          (global ADMINS or tenant admin)
DELETE /v1/tenants/{tenantId}                              DeleteTenant          (global ADMINS only)
POST   /v1/tenants/{tenantId}/members                      AddTenantMember       (global ADMINS or tenant admin)
DELETE /v1/tenants/{tenantId}/members/{memberId}           RemoveTenantMember    (global ADMINS or tenant admin)
```

### Role Definitions (tenant admin only)

```
GET    /v1/tenants/{tenantId}/roles                        ListRoleDefinitions
POST   /v1/tenants/{tenantId}/roles                        CreateRoleDefinition
PATCH  /v1/tenants/{tenantId}/roles/{roleId}               UpdateRoleDefinition
DELETE /v1/tenants/{tenantId}/roles/{roleId}               DeleteRoleDefinition
```

### Ownership Policies (tenant admin only)

```
GET    /v1/tenants/{tenantId}/ownership-policies                         ListOwnershipPolicies
PATCH  /v1/tenants/{tenantId}/ownership-policies/{resourceType}          UpdateOwnershipPolicy
```

### Team Creation under Tenant (tenant admin only)

```
POST   /v1/tenants/{tenantId}/teams                        CreateTeam (tenanted)
```

---

## Terraform User Stories

### US-1: DynamoDB — Tenant tables

> As infrastructure, I need `{prefix}-tenants` and `{prefix}-tenant-members` DynamoDB tables with appropriate GSIs so tenants and their members can be stored and queried efficiently.

Acceptance criteria:
- `tenants` table: hash key `id`, PAY_PER_REQUEST billing.
- `tenant-members` table: hash key `id`, GSIs: `tenantIdIndex` (hash: `tenantId`), `userIdIndex` (hash: `userId`), `tenantUserIndex` (hash: `tenantId`, range: `userId`).

### US-2: DynamoDB — Role definitions table

> As infrastructure, I need a `{prefix}-role-definitions` DynamoDB table so tenants can define custom permission sets for their roles.

Acceptance criteria:
- Hash key `id`, PAY_PER_REQUEST.
- GSI: `tenantIdIndex` (hash: `tenantId`).
- Global default records seeded via a one-time Lambda or Terraform local-exec: `admin`, `trainer`, `member` matching the agreed permission matrix.

### US-3: DynamoDB — Ownership policies table

> As infrastructure, I need a `{prefix}-ownership-policies` DynamoDB table so tenants can configure which permissions are granted to resource owners.

Acceptance criteria:
- Hash key `id`, PAY_PER_REQUEST.
- GSIs: `tenantIdIndex` (hash: `tenantId`), `tenantResourceTypeIndex` (hash: `tenantId`, range: `resourceType`).
- Global default records seeded for `goals`, `progress_reports`, `progress`, `comments`.

### US-4: DynamoDB — Extend existing tables

> As infrastructure, I need the existing `teams` table updated to support the tenant model.

Acceptance criteria:
- `teams` table gains `tenantId` attribute and GSI `tenantIdIndex` (hash: `tenantId`).
- No breaking changes to existing records; `tenantId` is nullable.

### US-5: Lambda IAM — Tenant management endpoints

> As infrastructure, I need IAM statements for the tenant management Lambda functions so they can read/write the new DynamoDB tables.

Acceptance criteria:
- `CreateTenant`, `UpdateTenant`, `DeleteTenant`: `dynamodb:PutItem`, `dynamodb:GetItem`, `dynamodb:UpdateItem`, `dynamodb:DeleteItem` on `tenants` table.
- `AddTenantMember`, `RemoveTenantMember`: `dynamodb:PutItem`, `dynamodb:UpdateItem`, `dynamodb:DeleteItem`, `dynamodb:Query` on `tenant-members` table and its GSIs.

### US-6: Lambda IAM — Role definition endpoints

> As infrastructure, I need IAM statements for role definition Lambda functions so tenant admins can manage permission sets.

Acceptance criteria:
- `ListRoleDefinitions`: `dynamodb:Query` on `role-definitions` GSI `tenantIdIndex`.
- `CreateRoleDefinition`: `dynamodb:PutItem` on `role-definitions`.
- `UpdateRoleDefinition`: `dynamodb:UpdateItem`, `dynamodb:GetItem` on `role-definitions`.
- `DeleteRoleDefinition`: `dynamodb:DeleteItem`, `dynamodb:GetItem` on `role-definitions`.

### US-7: Lambda IAM — Ownership policy endpoints

> As infrastructure, I need IAM statements for ownership policy Lambda functions.

Acceptance criteria:
- `ListOwnershipPolicies`: `dynamodb:Query` on `ownership-policies` GSI `tenantIdIndex`.
- `UpdateOwnershipPolicy`: `dynamodb:PutItem`, `dynamodb:Query` on `ownership-policies` table and GSI `tenantResourceTypeIndex`.

### US-8: Lambda IAM — Permission evaluation reads (all existing Lambdas)

> As infrastructure, all existing Lambda functions need read access to the new permission tables so the `CheckPermission` evaluator can resolve role definitions and ownership policies at runtime.

Acceptance criteria:
- All existing Lambda modules gain `dynamodb:GetItem`, `dynamodb:Query` on `role-definitions` and `ownership-policies` tables and all their GSIs.
- This is additive — no existing IAM statements are removed.

### US-9: API Gateway — Tenant, role, and ownership routes

> As infrastructure, I need API Gateway resources and methods wired up for all new endpoints with the Cognito authorizer attached.

Acceptance criteria:
- Resources created: `/v1/tenants`, `/v1/tenants/{tenantId}`, `/v1/tenants/{tenantId}/members`, `/v1/tenants/{tenantId}/members/{memberId}`, `/v1/tenants/{tenantId}/roles`, `/v1/tenants/{tenantId}/roles/{roleId}`, `/v1/tenants/{tenantId}/ownership-policies`, `/v1/tenants/{tenantId}/ownership-policies/{resourceType}`, `/v1/tenants/{tenantId}/teams`.
- All routes use the existing Cognito authorizer (`aws_api_gateway_authorizer.this`).
- CORS enabled on all new routes, matching `local.cors_allowed_origin`.
- Each route wired to its corresponding Lambda module using the existing `terraform-aws-microservice` module pattern.

---

## Out of Scope

- Global (`ADMINS`/`USERS`) Cognito-level permissions — unchanged.
- Tenant onboarding UI/self-service signup flow — future work.
- User-specific permission overrides — deliberately excluded to keep the model lean.
- Migration of existing teams to tenants — separate migration task.
