# Dynamic Permissions Model — Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace hard-coded team role checks with a dynamic RBAC + ownership policy engine backed by DynamoDB, and introduce a lightweight Tenant entity on the Team model.

**Architecture:** A single `CheckPermission(ctx, actorId, teamId, resource, action)` function evaluates ownership policies and tenant-scoped role definitions in order, replacing all existing `IsTeamAdmin/IsTeamTrainer/HasTeamAccess` calls. Global default role definitions and ownership policies are seeded at deploy time and apply to any team without a tenant-specific override.

**Tech Stack:** Go 1.21+, AWS Lambda + API Gateway, DynamoDB (aws-sdk-go-v2), Terraform, testify

> **Scope note:** This is Plan A of two. It covers the data model, DB layer, Terraform tables, permission engine, seed data, and migration of all existing handlers. Plan B (to be written separately) covers the new Tenant/Role/OwnershipPolicy management API endpoints.

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `models/permission.go` | Create | Permission string constants (`goals:read` etc.) + `Resource` struct |
| `models/tenant.go` | Create | `Tenant`, `TenantMember` structs |
| `models/role_definition.go` | Create | `RoleDefinition` struct |
| `models/ownership_policy.go` | Create | `OwnershipPolicy` struct |
| `models/teams.go` | Modify | Add `TenantId *string` field |
| `db/local_init.go` | Modify | Add table name vars for new tables |
| `db/lambda_init.go` | Modify | Add table name env vars for new tables |
| `db/tenants.go` | Create | DynamoDB CRUD for Tenant + TenantMember |
| `db/role_definitions.go` | Create | DynamoDB CRUD for RoleDefinition |
| `db/ownership_policies.go` | Create | DynamoDB CRUD for OwnershipPolicy |
| `utils/check_permission.go` | Create | `PermissionChecker` struct + `CheckPermission` function |
| `utils/check_permission_test.go` | Create | Table-driven unit tests for permission evaluation |
| `utils/permissions.go` | Modify | Add `HasTeamPermission` wrapper; keep old helpers as pass-throughs during migration |
| `scripts/seed_defaults/main.go` | Create | One-time seed script for global role definitions + ownership policies |
| `router/teams/teams_router.go` | Modify | Replace role checks with `HasTeamPermission` |
| `router/team-settings/team_settings_router.go` | Modify | Replace role checks |
| `router/team-members/team_member_router.go` | Modify | Replace role checks |
| `router/invites/invites_router.go` | Modify | Replace role checks |
| `router/seasons/seasons_router.go` | Modify | Replace role checks |
| `router/goals/goals_router.go` | Modify | Replace role checks + add `OwnerId` to resource |
| `router/progress-reports/progress_report_router.go` | Modify | Replace role checks + add `OwnerId` |
| `router/comments/comments_router.go` | Modify | Replace role checks + add `OwnerId` + `ParentOwnedBy` |
| `db.tf` | Modify | Add `tenantId` GSI to `teams` table |
| `tenants_infra.tf` | Create | New DynamoDB tables for tenants, role definitions, ownership policies |
| `main.tf` | Modify | Add read IAM statements for new tables to all existing Lambda modules |

---

## Task 1: Permission constants and new Go models

**Files:**
- Create: `files/src/models/permission.go`
- Create: `files/src/models/tenant.go`
- Create: `files/src/models/role_definition.go`
- Create: `files/src/models/ownership_policy.go`

- [ ] **Step 1: Create `models/permission.go`**

```go
package models

// Permission constants — resource:action pairs used in RoleDefinition.Permissions
// and OwnershipPolicy.OwnerPermissions.
const (
	PermTeamsRead   = "teams:read"
	PermTeamsWrite  = "teams:write"
	PermTeamsDelete = "teams:delete"

	PermTeamSettingsRead  = "team_settings:read"
	PermTeamSettingsWrite = "team_settings:write"

	PermMembersRead   = "members:read"
	PermMembersWrite  = "members:write"
	PermMembersDelete = "members:delete"

	PermInvitesRead   = "invites:read"
	PermInvitesWrite  = "invites:write"
	PermInvitesDelete = "invites:delete"

	PermSeasonsRead   = "seasons:read"
	PermSeasonsWrite  = "seasons:write"
	PermSeasonsDelete = "seasons:delete"

	PermGoalsRead   = "goals:read"
	PermGoalsWrite  = "goals:write"
	PermGoalsDelete = "goals:delete"

	PermProgressReportsRead   = "progress_reports:read"
	PermProgressReportsWrite  = "progress_reports:write"
	PermProgressReportsDelete = "progress_reports:delete"

	PermProgressRead  = "progress:read"
	PermProgressWrite = "progress:write"

	PermCommentsRead   = "comments:read"
	PermCommentsWrite  = "comments:write"
	PermCommentsDelete = "comments:delete"

	PermActivitiesRead = "activities:read"
)

// ResourceTypeGoals etc. are the resource type strings passed to CheckPermission.
const (
	ResourceTypeTeams          = "teams"
	ResourceTypeTeamSettings   = "team_settings"
	ResourceTypeMembers        = "members"
	ResourceTypeInvites        = "invites"
	ResourceTypeSeasons        = "seasons"
	ResourceTypeGoals          = "goals"
	ResourceTypeProgressReports = "progress_reports"
	ResourceTypeProgress       = "progress"
	ResourceTypeComments       = "comments"
	ResourceTypeActivities     = "activities"
)

// Resource describes the resource being accessed. OwnedBy is the direct owner
// (creator). ParentOwnedBy is the owner of the parent resource, used for
// comments where the parent goal/report owner also gets access.
type Resource struct {
	Type          string
	OwnedBy       string
	ParentOwnedBy string
}
```

- [ ] **Step 2: Create `models/tenant.go`**

```go
package models

import "time"

type TenantMemberRole   string
type TenantMemberStatus string

const (
	TenantMemberRoleAdmin  TenantMemberRole = "tenant_admin"
	TenantMemberRoleMember TenantMemberRole = "tenant_member"

	TenantMemberStatusActive  TenantMemberStatus = "active"
	TenantMemberStatusRemoved TenantMemberStatus = "removed"
)

type Tenant struct {
	Id        string    `dynamodbav:"id" json:"id"`
	Name      string    `dynamodbav:"name" json:"name"`
	OwnerId   string    `dynamodbav:"ownerId" json:"ownerId"`
	CreatedAt time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}

type TenantMember struct {
	Id        string             `dynamodbav:"id" json:"id"`
	TenantId  string             `dynamodbav:"tenantId" json:"tenantId"`
	UserId    string             `dynamodbav:"userId" json:"userId"`
	Role      TenantMemberRole   `dynamodbav:"role" json:"role"`
	Status    TenantMemberStatus `dynamodbav:"status" json:"status"`
	CreatedAt time.Time          `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `dynamodbav:"updatedAt" json:"updatedAt"`
}
```

- [ ] **Step 3: Create `models/role_definition.go`**

```go
package models

import "time"

type RoleDefinition struct {
	Id          string    `dynamodbav:"id" json:"id"`
	TenantId    string    `dynamodbav:"tenantId" json:"tenantId"` // "global" = applies to all tenants
	Name        string    `dynamodbav:"name" json:"name"`
	Permissions []string  `dynamodbav:"permissions" json:"permissions"`
	IsDefault   bool      `dynamodbav:"isDefault" json:"isDefault"`
	CreatedAt   time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}
```

> Note: `TenantId = "global"` is the sentinel for global defaults (DynamoDB cannot index null values).

- [ ] **Step 4: Create `models/ownership_policy.go`**

```go
package models

import "time"

type OwnershipPolicy struct {
	Id                     string    `dynamodbav:"id" json:"id"`
	TenantId               string    `dynamodbav:"tenantId" json:"tenantId"` // "global" = applies to all tenants
	ResourceType           string    `dynamodbav:"resourceType" json:"resourceType"`
	OwnerPermissions       []string  `dynamodbav:"ownerPermissions" json:"ownerPermissions"`
	ParentOwnerPermissions []string  `dynamodbav:"parentOwnerPermissions" json:"parentOwnerPermissions"`
	CreatedAt              time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt              time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}
```

- [ ] **Step 5: Build to verify no compile errors**

```bash
cd files/src && go build ./models/...
```

Expected: no output (success).

- [ ] **Step 6: Commit**

```bash
git add files/src/models/permission.go files/src/models/tenant.go files/src/models/role_definition.go files/src/models/ownership_policy.go
git commit -m "feat: add permission constants and new permission model types"
```

---

## Task 2: Extend Team model with TenantId

**Files:**
- Modify: `files/src/models/teams.go`

- [ ] **Step 1: Add `TenantId` field to `Team` struct**

In `files/src/models/teams.go`, add `TenantId *string` to the `Team` struct:

```go
type Team struct {
	Id        string     `dynamodbav:"id" json:"id"`
	Name      string     `dynamodbav:"teamName" json:"name"`
	Status    TeamStatus `dynamodbav:"status" json:"status"`
	Picture   string     `dynamodbav:"picture" json:"picture"`
	TenantId  *string    `dynamodbav:"tenantId,omitempty" json:"tenantId,omitempty"`
	CreatedAt time.Time  `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt time.Time  `dynamodbav:"updatedAt" json:"updatedAt"`
	DeletedAt *time.Time `dynamodbav:"deletedAt" json:"deletedAt"`
}
```

- [ ] **Step 2: Build to verify no compile errors**

```bash
cd files/src && go build ./models/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add files/src/models/teams.go
git commit -m "feat: add TenantId field to Team model"
```

---

## Task 3: DB layer — table name variables

**Files:**
- Modify: `files/src/db/local_init.go`
- Modify: `files/src/db/lambda_init.go`

- [ ] **Step 1: Add table name vars to `db/local_init.go`**

Add the four new table names to the existing `var` block in `files/src/db/local_init.go`:

```go
var (
	teamsTableName              = "dev-teams"
	invitesTableName            = "dev-invites"
	teamMembersTableName        = "dev-team-members"
	teamSettingsTableName       = "dev-team-settings"
	seasonsTableName            = "dev-seasons"
	goalsTableName              = "dev-goals"
	progressReportsTableName    = "dev-progress-reports"
	progressTableName           = "dev-progress"
	commentsTableName           = "dev-comments"
	commentFilesTableName       = "dev-comment-files"
	activitiesTableName         = "dev-activities"
	tenantsTableName            = "dev-tenants"
	tenantMembersTableName      = "dev-tenant-members"
	roleDefinitionsTableName    = "dev-role-definitions"
	ownershipPoliciesTableName  = "dev-ownership-policies"
)
```

- [ ] **Step 2: Add table name env vars to `db/lambda_init.go`**

Add the four new table names to the existing `var` block in `files/src/db/lambda_init.go`:

```go
var (
	teamsTableName              = os.Getenv("TEAMS_TABLE_NAME")
	invitesTableName            = os.Getenv("INVITE_TABLE_NAME")
	teamMembersTableName        = os.Getenv("TEAM_MEMBERS_TABLE_NAME")
	teamSettingsTableName       = os.Getenv("TEAM_SETTINGS_TABLE_NAME")
	seasonsTableName            = os.Getenv("SEASONS_TABLE_NAME")
	goalsTableName              = os.Getenv("GOALS_TABLE_NAME")
	progressReportsTableName    = os.Getenv("PROGRESS_REPORTS_TABLE_NAME")
	progressTableName           = os.Getenv("PROGRESS_TABLE_NAME")
	commentsTableName           = os.Getenv("COMMENTS_TABLE_NAME")
	commentFilesTableName       = os.Getenv("COMMENT_FILES_TABLE_NAME")
	activitiesTableName         = os.Getenv("ACTIVITIES_TABLE_NAME")
	tenantsTableName            = os.Getenv("TENANTS_TABLE_NAME")
	tenantMembersTableName      = os.Getenv("TENANT_MEMBERS_TABLE_NAME")
	roleDefinitionsTableName    = os.Getenv("ROLE_DEFINITIONS_TABLE_NAME")
	ownershipPoliciesTableName  = os.Getenv("OWNERSHIP_POLICIES_TABLE_NAME")
)
```

- [ ] **Step 3: Build to verify**

```bash
cd files/src && go build ./db/...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add files/src/db/local_init.go files/src/db/lambda_init.go
git commit -m "feat: add table name variables for permission model tables"
```

---

## Task 4: DB layer — tenants and tenant members

**Files:**
- Create: `files/src/db/tenants.go`

- [ ] **Step 1: Create `db/tenants.go`**

```go
package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

func CreateTenant(ctx context.Context, name, ownerId string) (*models.Tenant, error) {
	client = GetClient()
	now := time.Now()
	tenant := &models.Tenant{
		Id:        models.GenerateID(),
		Name:      name,
		OwnerId:   ownerId,
		CreatedAt: now,
		UpdatedAt: now,
	}
	item, err := attributevalue.MarshalMap(tenant)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tenantsTableName,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func GetTenantById(ctx context.Context, tenantId string) (*models.Tenant, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tenantsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: tenantId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var tenant models.Tenant
	err = attributevalue.UnmarshalMap(result.Item, &tenant)
	return &tenant, err
}

func DeleteTenantById(ctx context.Context, tenantId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &tenantsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: tenantId},
		},
	})
	return err
}

func AddTenantMember(ctx context.Context, tenantId, userId string, role models.TenantMemberRole) (*models.TenantMember, error) {
	client = GetClient()
	now := time.Now()
	member := &models.TenantMember{
		Id:        models.GenerateID(),
		TenantId:  tenantId,
		UserId:    userId,
		Role:      role,
		Status:    models.TenantMemberStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	item, err := attributevalue.MarshalMap(member)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tenantMembersTableName,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	return member, nil
}

func RemoveTenantMember(ctx context.Context, memberId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &tenantMembersTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: memberId},
		},
	})
	return err
}

func GetTenantMemberByUserAndTenant(ctx context.Context, userId, tenantId string) (*models.TenantMember, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &tenantMembersTableName,
		IndexName:              aws.String("tenantUserIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid AND userId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: tenantId},
			":uid": &types.AttributeValueMemberS{Value: userId},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var member models.TenantMember
	err = attributevalue.UnmarshalMap(result.Items[0], &member)
	return &member, err
}

func IsTenantAdmin(ctx context.Context, userId, tenantId string) (bool, error) {
	member, err := GetTenantMemberByUserAndTenant(ctx, userId, tenantId)
	if err != nil {
		return false, err
	}
	return member != nil && member.Role == models.TenantMemberRoleAdmin && member.Status == models.TenantMemberStatusActive, nil
}
```

- [ ] **Step 2: Build to verify**

```bash
cd files/src && go build ./db/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add files/src/db/tenants.go
git commit -m "feat: add DB layer for tenants and tenant members"
```

---

## Task 5: DB layer — role definitions and ownership policies

**Files:**
- Create: `files/src/db/role_definitions.go`
- Create: `files/src/db/ownership_policies.go`

- [ ] **Step 1: Create `db/role_definitions.go`**

```go
package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

// GetRoleDefinitionByTenantAndName returns the role definition for the given
// tenantId and role name. If no tenant-specific definition exists it falls back
// to the global default (tenantId = "global"). Returns nil if not found.
func GetRoleDefinitionByTenantAndName(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
	client = GetClient()
	candidates := []string{tenantId, "global"}
	for _, tid := range candidates {
		if tid == "" {
			continue
		}
		result, err := client.Query(ctx, &dynamodb.QueryInput{
			TableName:              &roleDefinitionsTableName,
			IndexName:              aws.String("tenantNameIndex"),
			KeyConditionExpression: aws.String("tenantId = :tid AND #name = :name"),
			ExpressionAttributeNames: map[string]string{
				"#name": "name",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":tid":  &types.AttributeValueMemberS{Value: tid},
				":name": &types.AttributeValueMemberS{Value: roleName},
			},
			Limit: aws.Int32(1),
		})
		if err != nil {
			return nil, err
		}
		if len(result.Items) > 0 {
			var def models.RoleDefinition
			if err := attributevalue.UnmarshalMap(result.Items[0], &def); err != nil {
				return nil, err
			}
			return &def, nil
		}
	}
	return nil, nil
}

func ListRoleDefinitionsByTenant(ctx context.Context, tenantId string) ([]*models.RoleDefinition, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &roleDefinitionsTableName,
		IndexName:              aws.String("tenantIdIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: tenantId},
		},
	})
	if err != nil {
		return nil, err
	}
	var defs []*models.RoleDefinition
	err = attributevalue.UnmarshalListOfMaps(result.Items, &defs)
	return defs, err
}

func CreateRoleDefinition(ctx context.Context, tenantId, name string, permissions []string, isDefault bool) (*models.RoleDefinition, error) {
	client = GetClient()
	now := time.Now()
	def := &models.RoleDefinition{
		Id:          models.GenerateID(),
		TenantId:    tenantId,
		Name:        name,
		Permissions: permissions,
		IsDefault:   isDefault,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	item, err := attributevalue.MarshalMap(def)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &roleDefinitionsTableName,
		Item:      item,
	})
	if err != nil {
		return nil, err
	}
	return def, nil
}

func UpdateRoleDefinitionPermissions(ctx context.Context, roleId string, permissions []string) (*models.RoleDefinition, error) {
	client = GetClient()
	permList, err := attributevalue.MarshalList(permissions)
	if err != nil {
		return nil, err
	}
	result, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: &roleDefinitionsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: roleId},
		},
		UpdateExpression: aws.String("SET permissions = :p, updatedAt = :u"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":p": &types.AttributeValueMemberL{Value: permList},
			":u": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		return nil, err
	}
	var def models.RoleDefinition
	err = attributevalue.UnmarshalMap(result.Attributes, &def)
	return &def, err
}

func DeleteRoleDefinition(ctx context.Context, roleId string) error {
	client = GetClient()
	_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &roleDefinitionsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: roleId},
		},
	})
	return err
}
```

- [ ] **Step 2: Create `db/ownership_policies.go`**

```go
package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/fpgschiba/volleygoals/models"
)

// GetOwnershipPolicy returns the ownership policy for the given tenantId and
// resourceType. Falls back to the global default (tenantId = "global").
func GetOwnershipPolicy(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error) {
	client = GetClient()
	candidates := []string{tenantId, "global"}
	for _, tid := range candidates {
		if tid == "" {
			continue
		}
		result, err := client.Query(ctx, &dynamodb.QueryInput{
			TableName:              &ownershipPoliciesTableName,
			IndexName:              aws.String("tenantResourceTypeIndex"),
			KeyConditionExpression: aws.String("tenantId = :tid AND resourceType = :rt"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":tid": &types.AttributeValueMemberS{Value: tid},
				":rt":  &types.AttributeValueMemberS{Value: resourceType},
			},
			Limit: aws.Int32(1),
		})
		if err != nil {
			return nil, err
		}
		if len(result.Items) > 0 {
			var policy models.OwnershipPolicy
			if err := attributevalue.UnmarshalMap(result.Items[0], &policy); err != nil {
				return nil, err
			}
			return &policy, nil
		}
	}
	return nil, nil
}

func ListOwnershipPoliciesByTenant(ctx context.Context, tenantId string) ([]*models.OwnershipPolicy, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &ownershipPoliciesTableName,
		IndexName:              aws.String("tenantIdIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: tenantId},
		},
	})
	if err != nil {
		return nil, err
	}
	var policies []*models.OwnershipPolicy
	err = attributevalue.UnmarshalListOfMaps(result.Items, &policies)
	return policies, err
}

func UpsertOwnershipPolicy(ctx context.Context, tenantId, resourceType string, ownerPerms, parentOwnerPerms []string) (*models.OwnershipPolicy, error) {
	client = GetClient()
	// Check if one already exists
	existing, err := GetOwnershipPolicyByTenantAndType(ctx, tenantId, resourceType)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if existing != nil {
		ownerList, err := attributevalue.MarshalList(ownerPerms)
		if err != nil {
			return nil, err
		}
		parentList, err := attributevalue.MarshalList(parentOwnerPerms)
		if err != nil {
			return nil, err
		}
		_, err = client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
			TableName: &ownershipPoliciesTableName,
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: existing.Id},
			},
			UpdateExpression: aws.String("SET ownerPermissions = :op, parentOwnerPermissions = :pp, updatedAt = :u"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":op": &types.AttributeValueMemberL{Value: ownerList},
				":pp": &types.AttributeValueMemberL{Value: parentList},
				":u":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			},
		})
		if err != nil {
			return nil, err
		}
		existing.OwnerPermissions = ownerPerms
		existing.ParentOwnerPermissions = parentOwnerPerms
		existing.UpdatedAt = now
		return existing, nil
	}
	policy := &models.OwnershipPolicy{
		Id:                     models.GenerateID(),
		TenantId:               tenantId,
		ResourceType:           resourceType,
		OwnerPermissions:       ownerPerms,
		ParentOwnerPermissions: parentOwnerPerms,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	item, err := attributevalue.MarshalMap(policy)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &ownershipPoliciesTableName,
		Item:      item,
	})
	return policy, err
}

// GetOwnershipPolicyByTenantAndType fetches the exact record for a given
// tenantId (no fallback). Used internally by UpsertOwnershipPolicy.
func GetOwnershipPolicyByTenantAndType(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error) {
	client = GetClient()
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &ownershipPoliciesTableName,
		IndexName:              aws.String("tenantResourceTypeIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid AND resourceType = :rt"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: tenantId},
			":rt":  &types.AttributeValueMemberS{Value: resourceType},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var policy models.OwnershipPolicy
	err = attributevalue.UnmarshalMap(result.Items[0], &policy)
	return &policy, err
}
```

- [ ] **Step 3: Build to verify**

```bash
cd files/src && go build ./db/...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
git add files/src/db/role_definitions.go files/src/db/ownership_policies.go
git commit -m "feat: add DB layer for role definitions and ownership policies"
```

---

## Task 6: Terraform — new DynamoDB tables and teams GSI

**Files:**
- Create: `tenants_infra.tf`
- Modify: `db.tf`

> Note: the `role-definitions` table needs a composite GSI `tenantNameIndex` (hash: `tenantId`, range: `name`) to support `GetRoleDefinitionByTenantAndName`. Add this in addition to the plain `tenantIdIndex`.

- [ ] **Step 1: Create `tenants_infra.tf`**

```hcl
# Tenants table
resource "aws_dynamodb_table" "tenants" {
  name         = "${var.prefix}-tenants"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

# Tenant members table
resource "aws_dynamodb_table" "tenant_members" {
  name         = "${var.prefix}-tenant-members"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "tenantId"
    type = "S"
  }
  attribute {
    name = "userId"
    type = "S"
  }

  global_secondary_index {
    name            = "tenantIdIndex"
    hash_key        = "tenantId"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "userIdIndex"
    hash_key        = "userId"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "tenantUserIndex"
    hash_key        = "tenantId"
    range_key       = "userId"
    projection_type = "ALL"
  }

  tags = local.tags
}

# Role definitions table
resource "aws_dynamodb_table" "role_definitions" {
  name         = "${var.prefix}-role-definitions"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "tenantId"
    type = "S"
  }
  attribute {
    name = "name"
    type = "S"
  }

  global_secondary_index {
    name            = "tenantIdIndex"
    hash_key        = "tenantId"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "tenantNameIndex"
    hash_key        = "tenantId"
    range_key       = "name"
    projection_type = "ALL"
  }

  tags = local.tags
}

# Ownership policies table
resource "aws_dynamodb_table" "ownership_policies" {
  name         = "${var.prefix}-ownership-policies"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "tenantId"
    type = "S"
  }
  attribute {
    name = "resourceType"
    type = "S"
  }

  global_secondary_index {
    name            = "tenantIdIndex"
    hash_key        = "tenantId"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "tenantResourceTypeIndex"
    hash_key        = "tenantId"
    range_key       = "resourceType"
    projection_type = "ALL"
  }

  tags = local.tags
}
```

- [ ] **Step 2: Add `tenantId` attribute and GSI to `teams` table in `db.tf`**

In `db.tf`, update the `aws_dynamodb_table.teams` resource to add:

```hcl
resource "aws_dynamodb_table" "teams" {
  name         = "${var.prefix}-teams"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "tenantId"
    type = "S"
  }

  global_secondary_index {
    name            = "tenantIdIndex"
    hash_key        = "tenantId"
    projection_type = "ALL"
  }

  tags = local.tags
}
```

- [ ] **Step 3: Add table name env vars to `locals.tf`**

Open `locals.tf` and find `lambda_environment_variables`. Add the four new table name references:

```hcl
TENANTS_TABLE_NAME           = aws_dynamodb_table.tenants.name
TENANT_MEMBERS_TABLE_NAME    = aws_dynamodb_table.tenant_members.name
ROLE_DEFINITIONS_TABLE_NAME  = aws_dynamodb_table.role_definitions.name
OWNERSHIP_POLICIES_TABLE_NAME = aws_dynamodb_table.ownership_policies.name
```

> Verify the exact variable name in `locals.tf` — look for the block that already contains `TEAMS_TABLE_NAME` etc.

- [ ] **Step 4: Validate Terraform**

```bash
cd D:/Projects/terraform-aws-modules/terraform-volleygoals && terraform validate
```

Expected: `Success! The configuration is valid.`

- [ ] **Step 5: Commit**

```bash
git add tenants_infra.tf db.tf locals.tf
git commit -m "feat: add Terraform DynamoDB tables for permission model + teams tenantId GSI"
```

---

## Task 7: Permission engine — CheckPermission with unit tests

**Files:**
- Create: `files/src/utils/check_permission.go`
- Create: `files/src/utils/check_permission_test.go`

- [ ] **Step 1: Write the failing tests first**

Create `files/src/utils/check_permission_test.go`:

```go
package utils_test

import (
	"context"
	"testing"

	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers — build a PermissionChecker with injected loaders.

func memberLoader(role string) func(ctx context.Context, userID, teamID string) (*models.TeamMember, error) {
	return func(ctx context.Context, userID, teamID string) (*models.TeamMember, error) {
		return &models.TeamMember{Role: models.TeamMemberRole(role)}, nil
	}
}

func teamLoader(tenantId *string) func(ctx context.Context, teamID string) (*models.Team, error) {
	return func(ctx context.Context, teamID string) (*models.Team, error) {
		return &models.Team{Id: teamID, TenantId: tenantId}, nil
	}
}

func ownershipLoader(ownerPerms, parentPerms []string) func(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error) {
	return func(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error) {
		return &models.OwnershipPolicy{
			OwnerPermissions:       ownerPerms,
			ParentOwnerPermissions: parentPerms,
		}, nil
	}
}

func roleLoader(perms []string) func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
	return func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
		return &models.RoleDefinition{Permissions: perms}, nil
	}
}

func nilRoleLoader() func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
	return func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
		return nil, nil
	}
}

func TestCheckPermission_OwnerCanReadOwnGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("member"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete}, nil),
		LoadRoleByTenant: nilRoleLoader(),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-1"}
	allowed, err := checker.Check(context.Background(), "user-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.True(t, allowed, "owner should be able to read their own goal")
}

func TestCheckPermission_NonOwnerMemberCannotReadGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("member"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead}, nil),
		LoadRoleByTenant: nilRoleLoader(),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-2"}
	allowed, err := checker.Check(context.Background(), "user-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.False(t, allowed, "non-owner member should not read another member's goal")
}

func TestCheckPermission_TrainerCanReadAnyGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("trainer"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader(nil, nil),
		LoadRoleByTenant: roleLoader([]string{models.PermGoalsRead, models.PermGoalsWrite}),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-2"}
	allowed, err := checker.Check(context.Background(), "trainer-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.True(t, allowed, "trainer role should be able to read any goal")
}

func TestCheckPermission_TrainerCannotEditTeam(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("trainer"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader(nil, nil),
		LoadRoleByTenant: roleLoader([]string{models.PermTeamsRead}), // trainer has read, not write
	}
	resource := models.Resource{Type: models.ResourceTypeTeams}
	allowed, err := checker.Check(context.Background(), "trainer-1", "team-1", resource, models.PermTeamsWrite)
	require.NoError(t, err)
	assert.False(t, allowed, "trainer should not be able to edit the team")
}

func TestCheckPermission_AdminCannotReadOtherMemberGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("admin"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead}, nil),
		LoadRoleByTenant: roleLoader([]string{models.PermTeamsRead, models.PermTeamsWrite, models.PermTeamsDelete}), // admin has no goals:read
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-other"}
	allowed, err := checker.Check(context.Background(), "admin-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.False(t, allowed, "admin should not read another member's goal")
}

func TestCheckPermission_AdminCanReadOwnGoal(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("admin"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete}, nil),
		LoadRoleByTenant: roleLoader([]string{models.PermTeamsWrite}),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "admin-1"}
	allowed, err := checker.Check(context.Background(), "admin-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.True(t, allowed, "admin should read their own goal via ownership")
}

func TestCheckPermission_GoalOwnerCanReadCommentViaParentOwnership(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("member"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermCommentsRead}, []string{models.PermCommentsRead, models.PermCommentsWrite}),
		LoadRoleByTenant: nilRoleLoader(),
	}
	// comment.OwnedBy = trainer, comment.ParentOwnedBy = member (goal owner)
	resource := models.Resource{
		Type:          models.ResourceTypeComments,
		OwnedBy:       "trainer-1",
		ParentOwnedBy: "member-1",
	}
	allowed, err := checker.Check(context.Background(), "member-1", "team-1", resource, models.PermCommentsRead)
	require.NoError(t, err)
	assert.True(t, allowed, "goal owner should read comments on their goal via parent ownership")
}

func TestCheckPermission_DenyByDefault(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember:   memberLoader("member"),
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader(nil, nil),
		LoadRoleByTenant: nilRoleLoader(),
	}
	resource := models.Resource{Type: models.ResourceTypeTeams}
	allowed, err := checker.Check(context.Background(), "member-1", "team-1", resource, models.PermTeamsDelete)
	require.NoError(t, err)
	assert.False(t, allowed, "should deny by default when no policy matches")
}

func TestCheckPermission_NilTeamMemberDenies(t *testing.T) {
	checker := &utils.PermissionChecker{
		LoadTeamMember: func(ctx context.Context, userID, teamID string) (*models.TeamMember, error) {
			return nil, nil // not a team member
		},
		LoadTeam:         teamLoader(nil),
		LoadOwnership:    ownershipLoader([]string{models.PermGoalsRead}, nil),
		LoadRoleByTenant: nilRoleLoader(),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals, OwnedBy: "user-1"}
	allowed, err := checker.Check(context.Background(), "user-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.False(t, allowed, "non-member should be denied even if they own the resource")
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd files/src && go test ./utils/... -run TestCheckPermission -v 2>&1 | head -20
```

Expected: compile error — `utils.PermissionChecker` and `checker.Check` are not defined yet.

- [ ] **Step 3: Create `utils/check_permission.go`**

```go
package utils

import (
	"context"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
)

// PermissionChecker evaluates whether an actor is allowed to perform an action
// on a resource within a team. Loader functions are injected for testability.
type PermissionChecker struct {
	LoadTeamMember   func(ctx context.Context, userID, teamID string) (*models.TeamMember, error)
	LoadTeam         func(ctx context.Context, teamID string) (*models.Team, error)
	LoadOwnership    func(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error)
	LoadRoleByTenant func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error)
}

// DefaultChecker wires the checker to the real db package.
var DefaultChecker = &PermissionChecker{
	LoadTeamMember:   db.GetTeamMemberByUserIDAndTeamID,
	LoadTeam:         db.GetTeamById,
	LoadOwnership:    db.GetOwnershipPolicy,
	LoadRoleByTenant: db.GetRoleDefinitionByTenantAndName,
}

// CheckPermission is the single entry point for all team-level permission
// checks. It replaces IsTeamAdmin, IsTeamTrainer, HasTeamAccess, etc.
func CheckPermission(ctx context.Context, actorId, teamId string, resource models.Resource, action string) (bool, error) {
	return DefaultChecker.Check(ctx, actorId, teamId, resource, action)
}

// Check runs the evaluation chain against the injected loaders.
func (pc *PermissionChecker) Check(ctx context.Context, actorId, teamId string, resource models.Resource, action string) (bool, error) {
	// Step 1: actor must be an active team member
	member, err := pc.LoadTeamMember(ctx, actorId, teamId)
	if err != nil {
		return false, err
	}
	if member == nil {
		return false, nil
	}

	// Resolve tenantId from team (empty string = no tenant = use global defaults only)
	team, err := pc.LoadTeam(ctx, teamId)
	if err != nil {
		return false, err
	}
	tenantId := ""
	if team != nil && team.TenantId != nil {
		tenantId = *team.TenantId
	}

	// Step 2: direct ownership check
	if resource.OwnedBy != "" && resource.OwnedBy == actorId {
		policy, err := pc.LoadOwnership(ctx, tenantId, resource.Type)
		if err != nil {
			return false, err
		}
		if policy != nil && containsString(policy.OwnerPermissions, action) {
			return true, nil
		}
	}

	// Step 3: parent ownership check
	if resource.ParentOwnedBy != "" && resource.ParentOwnedBy == actorId {
		policy, err := pc.LoadOwnership(ctx, tenantId, resource.Type)
		if err != nil {
			return false, err
		}
		if policy != nil && containsString(policy.ParentOwnerPermissions, action) {
			return true, nil
		}
	}

	// Step 4: tenant-specific role definition
	if tenantId != "" {
		roleDef, err := pc.LoadRoleByTenant(ctx, tenantId, string(member.Role))
		if err != nil {
			return false, err
		}
		if roleDef != nil && containsString(roleDef.Permissions, action) {
			return true, nil
		}
	}

	// Step 5: global default role definition
	roleDef, err := pc.LoadRoleByTenant(ctx, "global", string(member.Role))
	if err != nil {
		return false, err
	}
	if roleDef != nil && containsString(roleDef.Permissions, action) {
		return true, nil
	}

	return false, nil
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd files/src && go test ./utils/... -run TestCheckPermission -v
```

Expected: all 8 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add files/src/utils/check_permission.go files/src/utils/check_permission_test.go
git commit -m "feat: add CheckPermission evaluation engine with unit tests"
```

---

## Task 8: Add HasTeamPermission helper to utils/permissions.go

**Files:**
- Modify: `files/src/utils/permissions.go`

This helper bridges the new `CheckPermission` and the existing handler call sites, keeping migration diffs small.

- [ ] **Step 1: Add `HasTeamPermission` to `utils/permissions.go`**

Add the following function at the bottom of `files/src/utils/permissions.go`:

```go
// HasTeamPermission is a convenience wrapper around CheckPermission that
// extracts the actorId from the Cognito authorizer context.
// Use this in handlers to replace the old IsTeamAdmin / IsTeamTrainer / HasTeamAccess calls.
func HasTeamPermission(ctx context.Context, authorizer map[string]interface{}, teamId string, resource models.Resource, action string) bool {
	actorId := GetCognitoUsername(authorizer)
	if actorId == "" {
		return false
	}
	allowed, err := CheckPermission(ctx, actorId, teamId, resource, action)
	if err != nil {
		return false
	}
	return allowed
}
```

Add the required import for `models` at the top of the file if not already present.

- [ ] **Step 2: Build to verify**

```bash
cd files/src && go build ./utils/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add files/src/utils/permissions.go
git commit -m "feat: add HasTeamPermission wrapper for handler migration"
```

---

## Task 9: Seed script for global default data

**Files:**
- Create: `files/src/scripts/seed_defaults/main.go`

- [ ] **Step 1: Create the seed script**

```go
//go:build ignore

// seed_defaults seeds global RoleDefinition and OwnershipPolicy records into
// DynamoDB. Run once per environment:
//   go run files/src/scripts/seed_defaults/main.go
//
// Requires: LOCAL_AWS_PROFILE or AWS credentials in environment.
// Set TEAMS_TABLE_NAME etc. env vars or the script uses dev- prefixed names.

package main

import (
	"context"
	"log"
	"os"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
)

func main() {
	db.InitClient(nil)
	ctx := context.Background()

	if err := seedRoleDefinitions(ctx); err != nil {
		log.Fatalf("seed role definitions: %v", err)
	}
	if err := seedOwnershipPolicies(ctx); err != nil {
		log.Fatalf("seed ownership policies: %v", err)
	}
	log.Println("✅ Seed complete")
}

func seedRoleDefinitions(ctx context.Context) error {
	roles := []struct {
		name        string
		permissions []string
	}{
		{
			name: "admin",
			permissions: []string{
				models.PermTeamsRead, models.PermTeamsWrite, models.PermTeamsDelete,
				models.PermTeamSettingsRead, models.PermTeamSettingsWrite,
				models.PermMembersRead, models.PermMembersWrite, models.PermMembersDelete,
				models.PermInvitesRead, models.PermInvitesWrite, models.PermInvitesDelete,
				models.PermSeasonsRead,
				models.PermActivitiesRead,
			},
		},
		{
			name: "trainer",
			permissions: []string{
				models.PermTeamsRead,
				models.PermTeamSettingsRead,
				models.PermMembersRead,
				models.PermSeasonsRead, models.PermSeasonsWrite, models.PermSeasonsDelete,
				models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete,
				models.PermProgressReportsRead, models.PermProgressReportsWrite, models.PermProgressReportsDelete,
				models.PermProgressRead, models.PermProgressWrite,
				models.PermCommentsRead, models.PermCommentsWrite, models.PermCommentsDelete,
				models.PermActivitiesRead,
			},
		},
		{
			name: "member",
			permissions: []string{
				models.PermTeamsRead,
				models.PermMembersRead,
				models.PermSeasonsRead,
			},
		},
	}

	for _, r := range roles {
		existing, err := db.GetRoleDefinitionByTenantAndName(ctx, "global", r.name)
		if err != nil {
			return err
		}
		if existing != nil {
			log.Printf("  role %q already exists, skipping", r.name)
			continue
		}
		def, err := db.CreateRoleDefinition(ctx, "global", r.name, r.permissions, true)
		if err != nil {
			return err
		}
		log.Printf("  created role %q (%s)", def.Name, def.Id)
	}
	return nil
}

func seedOwnershipPolicies(ctx context.Context) error {
	policies := []struct {
		resourceType   string
		ownerPerms     []string
		parentOwnerPerms []string
	}{
		{
			resourceType: models.ResourceTypeGoals,
			ownerPerms: []string{
				models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete,
				models.PermCommentsRead, models.PermCommentsWrite,
			},
			parentOwnerPerms: nil,
		},
		{
			resourceType: models.ResourceTypeProgressReports,
			ownerPerms: []string{
				models.PermProgressReportsRead, models.PermProgressReportsWrite, models.PermProgressReportsDelete,
				models.PermCommentsRead, models.PermCommentsWrite,
			},
			parentOwnerPerms: nil,
		},
		{
			resourceType: models.ResourceTypeProgress,
			ownerPerms:   []string{models.PermProgressRead, models.PermProgressWrite},
			parentOwnerPerms: nil,
		},
		{
			resourceType: models.ResourceTypeComments,
			ownerPerms:   []string{models.PermCommentsRead, models.PermCommentsWrite, models.PermCommentsDelete},
			parentOwnerPerms: []string{models.PermCommentsRead, models.PermCommentsWrite},
		},
	}

	for _, p := range policies {
		policy, err := db.UpsertOwnershipPolicy(ctx, "global", p.resourceType, p.ownerPerms, p.parentOwnerPerms)
		if err != nil {
			return err
		}
		log.Printf("  upserted ownership policy for %q (%s)", policy.ResourceType, policy.Id)
	}
	return nil
}

// Run with: go run -tags local ./scripts/seed_defaults/main.go
// The -tags local flag uses db/local_init.go which has hardcoded dev-* table names.
```

- [ ] **Step 2: Build to verify the script compiles**

```bash
cd files/src && go build -tags local ./scripts/seed_defaults/main.go
```

Expected: no output. (The `-tags local` flag uses `db/local_init.go` with hardcoded `dev-*` table names, avoiding the need to set env vars manually.)

- [ ] **Step 3: Commit**

```bash
git add files/src/scripts/seed_defaults/main.go
git commit -m "feat: add seed script for global default role definitions and ownership policies"
```

---

## Task 10: Migrate existing handlers to CheckPermission

**Files:**
- Modify: `files/src/router/teams/teams_router.go`
- Modify: `files/src/router/team-settings/team_settings_router.go`
- Modify: `files/src/router/team-members/team_member_router.go`
- Modify: `files/src/router/invites/invites_router.go`
- Modify: `files/src/router/seasons/seasons_router.go`
- Modify: `files/src/router/goals/goals_router.go`
- Modify: `files/src/router/progress-reports/progress_report_router.go`
- Modify: `files/src/router/comments/comments_router.go`

### Migration reference table

| Old call | Replace with |
|---|---|
| `utils.IsTeamAdmin(ctx, auth, teamId)` | `utils.HasTeamPermission(ctx, auth, teamId, models.Resource{Type: models.ResourceTypeTeams}, models.PermTeamsWrite)` |
| `utils.IsTeamTrainer(ctx, auth, teamId)` | `utils.HasTeamPermission(ctx, auth, teamId, models.Resource{Type: models.ResourceTypeSeasons}, models.PermSeasonsWrite)` |
| `utils.IsTeamAdminOrTrainer(ctx, auth, teamId)` | context-dependent — see per-handler notes below |
| `utils.HasTeamAccess(ctx, auth, teamId)` | `utils.HasTeamPermission(ctx, auth, teamId, models.Resource{Type: models.ResourceTypeTeams}, models.PermTeamsRead)` |
| `utils.IsTeamUser(ctx, auth, teamId)` | `utils.HasTeamPermission(ctx, auth, teamId, models.Resource{Type: models.ResourceTypeTeams}, models.PermTeamsRead)` |

- [ ] **Step 1: Migrate `router/teams/teams_router.go`**

Open `files/src/router/teams/teams_router.go`. Apply these replacements:

In `UpdateTeam` — replace:
```go
if !utils.IsAdmin(event.RequestContext.Authorizer) {
```
with:
```go
if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeTeams}, models.PermTeamsWrite) {
```

In `ListTeams` — no change (Cognito `IsAdmin` check is correct, this is a system-level list).

In `GetTeam` — replace:
```go
if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
```
with:
```go
if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeTeams}, models.PermTeamsRead) {
```

In `DeleteTeam` — replace:
```go
if !utils.IsAdmin(event.RequestContext.Authorizer) {
```
with:
```go
if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeTeams}, models.PermTeamsDelete) {
```

In `CreateTeam` — no change (Cognito `IsAdmin` check is correct).

In `UploadTeamPicture` — replace:
```go
if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
```
with:
```go
if !utils.IsAdmin(event.RequestContext.Authorizer) && !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeTeams}, models.PermTeamsWrite) {
```

Add `"github.com/fpgschiba/volleygoals/models"` to imports if not present.

- [ ] **Step 2: Migrate `router/goals/goals_router.go`**

Read the file first:
```bash
cat files/src/router/goals/goals_router.go
```

For `CreateGoal` — replace the team-goal type check:
```go
if request.Type == models.GoalTypeTeam {
    if !utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
```
with:
```go
if request.Type == models.GoalTypeTeam {
    if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeGoals}, models.PermGoalsWrite) {
```

Replace the individual goal check:
```go
} else {
    if !utils.HasTeamAccess(ctx, event.RequestContext.Authorizer, teamId) {
```
with:
```go
} else {
    if !utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeGoals}, models.PermGoalsWrite) {
```

For the trainer-assign-owner check:
```go
if request.OwnerId != nil && utils.IsTeamAdminOrTrainer(ctx, event.RequestContext.Authorizer, teamId) {
```
replace with:
```go
if request.OwnerId != nil && utils.HasTeamPermission(ctx, event.RequestContext.Authorizer, teamId, models.Resource{Type: models.ResourceTypeGoals}, models.PermGoalsWrite) {
```

For `GetGoal` — the access check should use the goal's `OwnerId` to enable ownership:
```go
// After loading the goal from DB:
actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
if !utils.IsAdmin(event.RequestContext.Authorizer) {
    allowed, err := utils.CheckPermission(ctx, actorId, teamId,
        models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
        models.PermGoalsRead)
    if err != nil || !allowed {
        return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
    }
}
```

For `UpdateGoal` — after loading the goal from DB, replace any role check with:
```go
actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
if !utils.IsAdmin(event.RequestContext.Authorizer) {
    allowed, err := utils.CheckPermission(ctx, actorId, teamId,
        models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
        models.PermGoalsWrite)
    if err != nil || !allowed {
        return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
    }
}
```

For `DeleteGoal` — same pattern with `models.PermGoalsDelete`:
```go
actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
if !utils.IsAdmin(event.RequestContext.Authorizer) {
    allowed, err := utils.CheckPermission(ctx, actorId, teamId,
        models.Resource{Type: models.ResourceTypeGoals, OwnedBy: goal.OwnerId},
        models.PermGoalsDelete)
    if err != nil || !allowed {
        return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
    }
}
```

- [ ] **Step 3: Migrate `router/progress-reports/progress_report_router.go`**

Read the file:
```bash
cat files/src/router/progress-reports/progress_report_router.go
```

Pattern for all progress report handlers: load the report from DB first, then call `CheckPermission` with `OwnedBy: report.OwnerId`:
```go
actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
allowed, err := utils.CheckPermission(ctx, actorId, teamId,
    models.Resource{Type: models.ResourceTypeProgressReports, OwnedBy: report.OwnerId},
    models.PermProgressReportsRead) // or Write / Delete
if err != nil || !allowed {
    return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
}
```

- [ ] **Step 4: Migrate `router/comments/comments_router.go`**

Read the file:
```bash
cat files/src/router/comments/comments_router.go
```

Comments need both `OwnedBy` (commenter) and `ParentOwnedBy` (goal/report owner). Load the comment and its parent goal/report from DB, then:
```go
actorId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
allowed, err := utils.CheckPermission(ctx, actorId, teamId,
    models.Resource{
        Type:          models.ResourceTypeComments,
        OwnedBy:       comment.CreatedBy,
        ParentOwnedBy: parentGoal.OwnerId, // or parentReport.OwnerId
    },
    models.PermCommentsRead) // or Write / Delete
```

- [ ] **Step 5: Migrate remaining routers**

Repeat for `team-settings`, `team-members`, `invites`, `seasons` using the reference table from the top of this task. Use `models.PermTeamSettingsWrite`, `models.PermMembersWrite`, `models.PermInvitesWrite`, `models.PermSeasonsWrite` respectively for write operations, and the corresponding `:read`/`:delete` for those operations.

- [ ] **Step 6: Build to verify**

```bash
cd files/src && go build ./...
```

Expected: no output.

- [ ] **Step 7: Run all tests**

```bash
cd files/src && go test ./... -v
```

Expected: all tests pass (including the 8 `TestCheckPermission_*` tests).

- [ ] **Step 8: Commit**

```bash
git add files/src/router/
git commit -m "feat: migrate all handlers to CheckPermission dynamic permission model"
```

---

## Task 11: Terraform — IAM for all existing Lambdas (US-8)

**Files:**
- Modify: `main.tf` (and all other `routes_*.tf` files that declare Lambda modules)

All existing Lambda modules need `dynamodb:GetItem` and `dynamodb:Query` on the four new tables and their GSIs so the `CheckPermission` evaluator can read role definitions and ownership policies at runtime.

- [ ] **Step 1: Add IAM statements to every existing Lambda module**

For each Lambda module in `main.tf` and `routes_*.tf`, add the following to `additional_iam_statements`:

```hcl
{
  actions = ["dynamodb:GetItem", "dynamodb:Query"]
  resources = [
    aws_dynamodb_table.role_definitions.arn,
    "${aws_dynamodb_table.role_definitions.arn}/index/tenantIdIndex",
    "${aws_dynamodb_table.role_definitions.arn}/index/tenantNameIndex",
    aws_dynamodb_table.ownership_policies.arn,
    "${aws_dynamodb_table.ownership_policies.arn}/index/tenantIdIndex",
    "${aws_dynamodb_table.ownership_policies.arn}/index/tenantResourceTypeIndex",
    aws_dynamodb_table.teams.arn,
  ]
},
```

> Note: `teams.arn` is already in the read statements for most modules. Skip the duplicate if it's already present and just add the role_definitions and ownership_policies lines.

- [ ] **Step 2: Validate Terraform**

```bash
cd D:/Projects/terraform-aws-modules/terraform-volleygoals && terraform validate
```

Expected: `Success! The configuration is valid.`

- [ ] **Step 3: Commit**

```bash
git add main.tf routes_*.tf
git commit -m "feat: add permission table read IAM to all existing Lambda modules"
```

---

## Done — Plan A Complete

When all tasks are complete:
- All existing API endpoints enforce dynamic RBAC + ownership permissions
- Global default role definitions and ownership policies are seeded
- Terraform provisions all new DynamoDB tables with correct GSIs
- All existing Lambdas can read the permission tables at runtime
- 8 unit tests cover the core permission evaluation logic

**Next:** Write and implement Plan B — Tenant Management API (handlers + Terraform routes for `/v1/tenants`, `/v1/tenants/{tenantId}/roles`, `/v1/tenants/{tenantId}/ownership-policies`).
