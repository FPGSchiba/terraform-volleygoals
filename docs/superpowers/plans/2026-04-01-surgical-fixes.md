# Surgical Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix 5 independent bugs covering IAM permissions, tenant ownership, duplicate memberships, redundant DB queries, and a deprecated Lambda runtime.

**Architecture:** All changes are isolated to specific files with no cross-cutting dependencies. The permission checker fix is the only change that requires new Go tests. Terraform changes are verified with `terraform plan`.

**Tech Stack:** Go 1.x, AWS DynamoDB (aws-sdk-go-v2), Terraform HCL, testify/assert

---

### Task 1: Fix seed Lambda IAM policy — add `dynamodb:UpdateItem`

**Files:**
- Modify: `seed_defaults.tf`

- [ ] **Step 1: Add `dynamodb:UpdateItem` to the IAM policy**

In `seed_defaults.tf`, update `aws_iam_role_policy.seed_defaults_dynamo` so the `Action` list includes `UpdateItem`:

```hcl
resource "aws_iam_role_policy" "seed_defaults_dynamo" {
  name = "${var.prefix}-seed-defaults-dynamo"
  role = aws_iam_role.seed_defaults.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:Query",
          "dynamodb:UpdateItem",
        ]
        Resource = [
          aws_dynamodb_table.role_definitions.arn,
          "${aws_dynamodb_table.role_definitions.arn}/index/tenantIdIndex",
          "${aws_dynamodb_table.role_definitions.arn}/index/tenantNameIndex",
          aws_dynamodb_table.ownership_policies.arn,
          "${aws_dynamodb_table.ownership_policies.arn}/index/tenantIdIndex",
          "${aws_dynamodb_table.ownership_policies.arn}/index/tenantResourceTypeIndex",
        ]
      }
    ]
  })
}
```

- [ ] **Step 2: Verify plan**

```bash
cd /path/to/terraform-volleygoals
terraform plan
```
Expected: plan shows update to `aws_iam_role_policy.seed_defaults_dynamo` with `dynamodb:UpdateItem` added.

- [ ] **Step 3: Commit**

```bash
git add seed_defaults.tf
git commit -m "fix: add dynamodb:UpdateItem to seed Lambda IAM policy"
```

---

### Task 2: `CreateTenant` — auto-create owner as `tenant_admin`

**Files:**
- Modify: `files/src/db/tenants.go`

- [ ] **Step 1: Update `CreateTenant` to call `AddTenantMember` after persisting the tenant**

In `files/src/db/tenants.go`, update the `CreateTenant` function:

```go
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
		return nil, fmt.Errorf("CreateTenant: marshal tenant: %w", err)
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tenantsTableName,
		Item:      item,
	})
	if err != nil {
		return nil, fmt.Errorf("CreateTenant: put tenant: %w", err)
	}
	if _, err = AddTenantMember(ctx, tenant.Id, ownerId, models.TenantMemberRoleAdmin); err != nil {
		return nil, fmt.Errorf("CreateTenant: add owner as admin: %w", err)
	}
	return tenant, nil
}
```

Add `"fmt"` to the import block if not already present.

- [ ] **Step 2: Build to verify no compile errors**

```bash
cd files/src
go build ./...
```
Expected: exits 0, no output.

- [ ] **Step 3: Commit**

```bash
git add files/src/db/tenants.go
git commit -m "fix: CreateTenant now auto-creates owner as tenant_admin member"
```

---

### Task 3: `AddTenantMember` — deterministic composite key

**Files:**
- Modify: `files/src/db/tenants.go`

- [ ] **Step 1: Replace `models.GenerateID()` with a deterministic composite key**

In `files/src/db/tenants.go`, update `AddTenantMember`:

```go
func AddTenantMember(ctx context.Context, tenantId, userId string, role models.TenantMemberRole) (*models.TenantMember, error) {
	client = GetClient()
	now := time.Now()
	member := &models.TenantMember{
		Id:        tenantId + "#" + userId,
		TenantId:  tenantId,
		UserId:    userId,
		Role:      role,
		Status:    models.TenantMemberStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	item, err := attributevalue.MarshalMap(member)
	if err != nil {
		return nil, fmt.Errorf("AddTenantMember: marshal: %w", err)
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tenantMembersTableName,
		Item:      item,
	})
	if err != nil {
		return nil, fmt.Errorf("AddTenantMember: put: %w", err)
	}
	return member, nil
}
```

`PutItem` with the same composite key is idempotent — re-adding an existing active member safely overwrites with the same data.

- [ ] **Step 2: Build to verify no compile errors**

```bash
cd files/src
go build ./...
```
Expected: exits 0, no output.

- [ ] **Step 3: Commit**

```bash
git add files/src/db/tenants.go
git commit -m "fix: use deterministic tenantId#userId key in AddTenantMember to prevent duplicates"
```

---

### Task 4: `PermissionChecker` — eliminate redundant role lookup

**Files:**
- Modify: `files/src/db/role_definitions.go`
- Modify: `files/src/utils/check_permission.go`
- Modify: `files/src/utils/check_permission_test.go`

- [ ] **Step 1: Write failing tests for the two-field checker**

Add to `files/src/utils/check_permission_test.go`:

```go
// exactRoleLoader returns a role only when tenantId matches exactly; never falls back to "global".
func exactRoleLoader(tenantId string, perms []string) func(ctx context.Context, tid, roleName string) (*models.RoleDefinition, error) {
	return func(ctx context.Context, tid, roleName string) (*models.RoleDefinition, error) {
		if tid == tenantId {
			return &models.RoleDefinition{Permissions: perms}, nil
		}
		return nil, nil
	}
}

func TestCheckPermission_TenantRoleUsedBeforeGlobal(t *testing.T) {
	tenantId := "tenant-abc"
	// Tenant role has goals:write; global role has only goals:read.
	checker := &utils.PermissionChecker{
		LoadTeamMember:        memberLoader("trainer"),
		LoadTeam:              teamLoader(&tenantId),
		LoadOwnership:         ownershipLoader(nil, nil),
		LoadRoleByTenantExact: exactRoleLoader(tenantId, []string{models.PermGoalsWrite}),
		LoadRoleByTenant:      roleLoader([]string{models.PermGoalsRead}),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals}
	allowed, err := checker.Check(context.Background(), "trainer-1", "team-1", resource, models.PermGoalsWrite)
	require.NoError(t, err)
	assert.True(t, allowed, "tenant-specific role should grant goals:write")
}

func TestCheckPermission_GlobalRoleFallbackWhenNoTenantRole(t *testing.T) {
	tenantId := "tenant-abc"
	// No tenant-specific role; global role has goals:read.
	checker := &utils.PermissionChecker{
		LoadTeamMember:        memberLoader("trainer"),
		LoadTeam:              teamLoader(&tenantId),
		LoadOwnership:         ownershipLoader(nil, nil),
		LoadRoleByTenantExact: exactRoleLoader("other-tenant", []string{models.PermGoalsWrite}), // won't match
		LoadRoleByTenant:      roleLoader([]string{models.PermGoalsRead}),
	}
	resource := models.Resource{Type: models.ResourceTypeGoals}
	allowed, err := checker.Check(context.Background(), "trainer-1", "team-1", resource, models.PermGoalsRead)
	require.NoError(t, err)
	assert.True(t, allowed, "global role fallback should grant goals:read")
}
```

- [ ] **Step 2: Run tests — expect compile failure (field not defined yet)**

```bash
cd files/src
go test ./utils/... -run TestCheckPermission_TenantRole -v
```
Expected: compile error — `unknown field 'LoadRoleByTenantExact' in struct literal`.

- [ ] **Step 3: Add `GetRoleDefinitionByTenantExact` to `db/role_definitions.go`**

Add after `GetRoleDefinitionByTenantAndName` in `files/src/db/role_definitions.go`:

```go
// GetRoleDefinitionByTenantExact returns the role definition for the given
// tenantId and role name with NO global fallback. Returns nil if not found.
func GetRoleDefinitionByTenantExact(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error) {
	client = GetClient()
	if tenantId == "" {
		return nil, nil
	}
	result, err := client.Query(ctx, &dynamodb.QueryInput{
		TableName:              &roleDefinitionsTableName,
		IndexName:              aws.String("tenantNameIndex"),
		KeyConditionExpression: aws.String("tenantId = :tid AND #name = :name"),
		ExpressionAttributeNames: map[string]string{
			"#name": "name",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid":  &types.AttributeValueMemberS{Value: tenantId},
			":name": &types.AttributeValueMemberS{Value: roleName},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("GetRoleDefinitionByTenantExact: query: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	var def models.RoleDefinition
	if err := attributevalue.UnmarshalMap(result.Items[0], &def); err != nil {
		return nil, fmt.Errorf("GetRoleDefinitionByTenantExact: unmarshal: %w", err)
	}
	return &def, nil
}
```

Add `"fmt"` to imports in `role_definitions.go` if not already present.

- [ ] **Step 4: Add `LoadRoleByTenantExact` field to `PermissionChecker` and update `Check`**

Replace the full content of `files/src/utils/check_permission.go`:

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
	LoadTeamMember        func(ctx context.Context, userID, teamID string) (*models.TeamMember, error)
	LoadTeam              func(ctx context.Context, teamID string) (*models.Team, error)
	LoadOwnership         func(ctx context.Context, tenantId, resourceType string) (*models.OwnershipPolicy, error)
	LoadRoleByTenantExact func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error)
	LoadRoleByTenant      func(ctx context.Context, tenantId, roleName string) (*models.RoleDefinition, error)
}

// DefaultChecker wires the checker to the real db package.
var DefaultChecker = &PermissionChecker{
	LoadTeamMember:        db.GetTeamMemberByUserIDAndTeamID,
	LoadTeam:              db.GetTeamById,
	LoadOwnership:         db.GetOwnershipPolicy,
	LoadRoleByTenantExact: db.GetRoleDefinitionByTenantExact,
	LoadRoleByTenant:      db.GetRoleDefinitionByTenantAndName,
}

// CheckPermission is the single entry point for all team-level permission checks.
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

	// Step 4: tenant-specific role definition (exact match only, no global fallback)
	if tenantId != "" {
		roleDef, err := pc.LoadRoleByTenantExact(ctx, tenantId, string(member.Role))
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

- [ ] **Step 5: Update existing tests — add `LoadRoleByTenantExact` field**

All existing `PermissionChecker` literals in `check_permission_test.go` need the new field. For tests that don't exercise tenant-specific roles, use `nilRoleLoader()`:

Replace every existing `&utils.PermissionChecker{` block that doesn't have `LoadRoleByTenantExact` to add:
```go
LoadRoleByTenantExact: nilRoleLoader(),
```

For example, `TestCheckPermission_OwnerCanReadOwnGoal` becomes:
```go
checker := &utils.PermissionChecker{
    LoadTeamMember:        memberLoader("member"),
    LoadTeam:              teamLoader(nil),
    LoadOwnership:         ownershipLoader([]string{models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete}, nil),
    LoadRoleByTenantExact: nilRoleLoader(),
    LoadRoleByTenant:      nilRoleLoader(),
}
```

Apply this same pattern to all 8 existing test cases (add `LoadRoleByTenantExact: nilRoleLoader()`).

- [ ] **Step 6: Run all permission tests — expect PASS**

```bash
cd files/src
go test ./utils/... -v
```
Expected: all 10 tests pass (8 existing + 2 new).

- [ ] **Step 7: Commit**

```bash
git add files/src/db/role_definitions.go files/src/utils/check_permission.go files/src/utils/check_permission_test.go
git commit -m "fix: split PermissionChecker role lookup into exact-tenant and global-fallback steps"
```

---

### Task 5: Lambda runtime — single `local.lambda_runtime` variable

**Files:**
- Modify: `locals.tf`
- Modify: `seed_defaults.tf`
- Modify: `routes_seasons.tf`, `routes_comments.tf`, `routes_invites.tf`, `routes_search_health.tf`, `routes_self.tf`, `routes_tenants.tf`, `routes_users.tf`
- Modify (external module): `github.com/FPGSchiba/terraform-aws-microservice` — `locals.tf`

- [ ] **Step 1: Add `lambda_runtime` to `locals.tf`**

In `locals.tf`, add inside the `locals {}` block after `lambda_layer_arns`:

```hcl
  lambda_runtime = "provided.al2023"
```

- [ ] **Step 2: Pass `runtime = local.lambda_runtime` in `seed_defaults.tf`**

In `seed_defaults.tf`, update `aws_lambda_function.seed_defaults`:

```hcl
resource "aws_lambda_function" "seed_defaults" {
  function_name    = "${var.prefix}-seed-defaults"
  role             = aws_iam_role.seed_defaults.arn
  filename         = data.archive_file.shared_lambda_zip.output_path
  source_code_hash = data.archive_file.shared_lambda_zip.output_base64sha256
  handler          = "bootstrap"
  runtime          = local.lambda_runtime
  timeout          = 60
  # ... rest unchanged
}
```

- [ ] **Step 3: Add `runtime = local.lambda_runtime` to every module call in route files**

Each of the following files contains module blocks that call `github.com/FPGSchiba/terraform-aws-microservice`. Add `runtime = local.lambda_runtime` as a parameter to **every** `module` block in each file:

- `routes_seasons.tf` — 17 module blocks
- `routes_tenants.tf` — 13 module blocks
- `routes_comments.tf` — 6 module blocks
- `routes_invites.tf` — 5 module blocks
- `routes_users.tf` — 4 module blocks
- `routes_search_health.tf` — 2 module blocks
- `routes_self.tf` — 1 module block

Example — in `routes_seasons.tf`, `module "create_season_ms"`:
```hcl
module "create_season_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  # ... existing params ...
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
}
```

Add `runtime = local.lambda_runtime` to every module block following this same pattern.

- [ ] **Step 4: Update external module `is_go_build_lambda` condition**

In the `github.com/FPGSchiba/terraform-aws-microservice` repository, update `locals.tf`:

```hcl
is_go_build_lambda = contains(["provided.al2", "provided.al2023"], var.runtime) && var.handler == null
```

Tag a new module version (e.g. `v2.4.2`) and update the `source` ref in all module blocks from `?ref=v2.4.1` to `?ref=v2.4.2`. Run `terraform init` to pull the new version.

> **Note:** If all module calls in this project already pass `pre_built_zip`, the `is_go_build_lambda` condition never triggers. This step is good hygiene for future use but can be deferred.

- [ ] **Step 5: Verify plan**

```bash
terraform plan
```
Expected: plan shows runtime updates for all Lambda functions from `provided.al2` → `provided.al2023`. No other resource changes.

- [ ] **Step 6: Commit**

```bash
git add locals.tf seed_defaults.tf routes_seasons.tf routes_comments.tf routes_invites.tf routes_search_health.tf routes_self.tf routes_tenants.tf routes_users.tf
git commit -m "fix: centralise Lambda runtime in locals.tf and upgrade to provided.al2023"
```
