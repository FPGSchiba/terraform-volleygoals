# Dynamic Permissions Model — Management API (Plan B)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose the dynamic permissions model via REST API — Tenant CRUD, Role Definition management, Ownership Policy management, tenanted Team creation, and automated seeding of global defaults via a Terraform-controlled Lambda invocation.

**Architecture:** All new handlers follow the single-binary Lambda pattern (shared `bootstrap` binary, handler selected via `HANDLER` env var). The seed Lambda reuses the same binary with a `SeedDefaults` handler, invoked by `aws_lambda_invocation` after the DynamoDB tables exist. Tenant-scoped operations guard with `utils.IsTenantAdmin` (DB lookup) or `utils.IsAdmin` (Cognito group). All new API routes live in `routes_tenants.tf` and use the existing `terraform-aws-microservice` module pattern.

**Tech Stack:** Go 1.21+, AWS Lambda (provided.al2, shared binary), API Gateway, DynamoDB (aws-sdk-go-v2), Terraform ≥ 1.10, terraform-aws-microservice v2.4.1, testify

> **Base branch:** `feature/dynamic-permissions-model` (Plan A complete)

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `files/src/db/tenants.go` | Modify | Add `UpdateTenant`, `GetTenantMemberById` |
| `files/src/db/role_definitions.go` | Modify | Add `GetRoleDefinitionById` |
| `files/src/router/seed/seed_router.go` | Create | `SeedDefaults` Lambda handler |
| `files/src/router/tenants/tenants_router.go` | Create | CreateTenant, GetTenant, UpdateTenant, DeleteTenant, AddTenantMember, RemoveTenantMember |
| `files/src/router/tenants/roles_router.go` | Create | ListRoleDefinitions, CreateRoleDefinition, UpdateRoleDefinition, DeleteRoleDefinition |
| `files/src/router/tenants/ownership_policies_router.go` | Create | ListOwnershipPolicies, UpdateOwnershipPolicy |
| `files/src/router/tenants/tenanted_teams_router.go` | Create | CreateTenantedTeam |
| `files/src/router/router.go` | Modify | Register all 14 new handler names in switch |
| `files/src/utils/http.go` | Modify | Add tenant-related error message constants |
| `seed_defaults.tf` | Create | IAM role + Lambda function + `aws_lambda_invocation` for seed |
| `routes_tenants.tf` | Create | API Gateway resources + 13 Lambda modules for all tenant routes |

---

## Task 1: Missing DB helpers

**Files:**
- Modify: `files/src/db/tenants.go`
- Modify: `files/src/db/role_definitions.go`

Three functions are needed by the management handlers that were not included in Plan A.

- [ ] **Step 1: Add `UpdateTenant` to `db/tenants.go`**

Open `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\db\tenants.go` and append:

```go
func UpdateTenant(ctx context.Context, tenant *models.Tenant) error {
	client = GetClient()
	tenant.UpdatedAt = time.Now()
	item, err := attributevalue.MarshalMap(tenant)
	if err != nil {
		return err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &tenantsTableName,
		Item:      item,
	})
	return err
}
```

- [ ] **Step 2: Add `GetTenantMemberById` to `db/tenants.go`**

Append after `UpdateTenant`:

```go
func GetTenantMemberById(ctx context.Context, memberId string) (*models.TenantMember, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &tenantMembersTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: memberId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var member models.TenantMember
	err = attributevalue.UnmarshalMap(result.Item, &member)
	return &member, err
}
```

- [ ] **Step 3: Add `GetRoleDefinitionById` to `db/role_definitions.go`**

Open `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\db\role_definitions.go` and append:

```go
func GetRoleDefinitionById(ctx context.Context, roleId string) (*models.RoleDefinition, error) {
	client = GetClient()
	result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &roleDefinitionsTableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: roleId},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var def models.RoleDefinition
	err = attributevalue.UnmarshalMap(result.Item, &def)
	return &def, err
}
```

- [ ] **Step 4: Build to verify**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go build ./db/...
```

Expected: no output.

- [ ] **Step 5: Commit**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals
git add files/src/db/tenants.go files/src/db/role_definitions.go
git commit -m "feat: add missing DB helpers for management API (UpdateTenant, GetTenantMemberById, GetRoleDefinitionById)"
```

---

## Task 2: Error message constants

**Files:**
- Modify: `files/src/utils/http.go`

- [ ] **Step 1: Add tenant-related error constants to `utils/http.go`**

Open `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\utils\http.go`. Find the last error constant block and add after it (before the `MsgSuccess` line):

```go
	// Tenant related errors
	MsgErrorTenantNotFound       ResponseMessage = "error.tenant.notFound"
	MsgErrorTenantMemberNotFound ResponseMessage = "error.tenant.memberNotFound"
	MsgErrorNotTenantAdmin       ResponseMessage = "error.tenant.notAdmin"

	// Role definition errors
	MsgErrorRoleNotFound      ResponseMessage = "error.role.notFound"
	MsgErrorRoleIsDefault     ResponseMessage = "error.role.isDefault"

	// Ownership policy errors
	MsgErrorOwnershipPolicyNotFound ResponseMessage = "error.ownershipPolicy.notFound"
```

- [ ] **Step 2: Build to verify**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go build ./utils/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals
git add files/src/utils/http.go
git commit -m "feat: add tenant-related error message constants"
```

---

## Task 3: SeedDefaults Lambda handler

**Files:**
- Create: `files/src/router/seed/seed_router.go`
- Modify: `files/src/router/router.go`

The seed script in `scripts/seed_defaults/main.go` is a standalone tool. This task wraps its logic in a proper Lambda handler so the shared binary can execute it when invoked by Terraform.

- [ ] **Step 1: Create `files/src/router/seed/seed_router.go`**

Create the directory and file at `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\router\seed\seed_router.go`:

```go
package seed

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

// SeedDefaults seeds global RoleDefinition and OwnershipPolicy records.
// Invoked by Terraform via aws_lambda_invocation after tables are created.
// Safe to re-run: existing records are skipped.
func SeedDefaults(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if err := seedRoleDefinitions(ctx); err != nil {
		log.WithError(err).Error("seed role definitions failed")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if err := seedOwnershipPolicies(ctx); err != nil {
		log.WithError(err).Error("seed ownership policies failed")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]string{"status": "seeded"})
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
			log.Infof("role %q already exists, skipping", r.name)
			continue
		}
		if _, err := db.CreateRoleDefinition(ctx, "global", r.name, r.permissions, true); err != nil {
			return err
		}
		log.Infof("created role %q", r.name)
	}
	return nil
}

func seedOwnershipPolicies(ctx context.Context) error {
	policies := []struct {
		resourceType     string
		ownerPerms       []string
		parentOwnerPerms []string
	}{
		{
			resourceType: models.ResourceTypeGoals,
			ownerPerms: []string{
				models.PermGoalsRead, models.PermGoalsWrite, models.PermGoalsDelete,
				models.PermCommentsRead, models.PermCommentsWrite,
			},
		},
		{
			resourceType: models.ResourceTypeProgressReports,
			ownerPerms: []string{
				models.PermProgressReportsRead, models.PermProgressReportsWrite, models.PermProgressReportsDelete,
				models.PermCommentsRead, models.PermCommentsWrite,
			},
		},
		{
			resourceType: models.ResourceTypeProgress,
			ownerPerms:   []string{models.PermProgressRead, models.PermProgressWrite},
		},
		{
			resourceType:     models.ResourceTypeComments,
			ownerPerms:       []string{models.PermCommentsRead, models.PermCommentsWrite, models.PermCommentsDelete},
			parentOwnerPerms: []string{models.PermCommentsRead, models.PermCommentsWrite},
		},
	}

	for _, p := range policies {
		if _, err := db.UpsertOwnershipPolicy(ctx, "global", p.resourceType, p.ownerPerms, p.parentOwnerPerms); err != nil {
			return err
		}
		log.Infof("upserted ownership policy for %q", p.resourceType)
	}
	return nil
}
```

- [ ] **Step 2: Register `SeedDefaults` in `router/router.go`**

Open `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\router\router.go`.

Add the import:
```go
"github.com/fpgschiba/volleygoals/router/seed"
```

Add the case to the switch statement (after the last existing case, before `default`):
```go
	// Seed handlers
	case "SeedDefaults":
		response, err = seed.SeedDefaults(ctx, event)
```

- [ ] **Step 3: Build to verify**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go build ./...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals
git add files/src/router/seed/seed_router.go files/src/router/router.go
git commit -m "feat: add SeedDefaults Lambda handler for Terraform-controlled seeding"
```

---

## Task 4: Terraform — seed Lambda invocation

**Files:**
- Create: `seed_defaults.tf`

This creates a standalone Lambda (no API Gateway) that runs the `SeedDefaults` handler once after the permission tables exist. Re-seeding is controlled by bumping `seed_version`.

- [ ] **Step 1: Create `seed_defaults.tf`**

Create `D:\Projects\terraform-aws-modules\terraform-volleygoals\seed_defaults.tf`:

```hcl
# IAM role for seed Lambda
resource "aws_iam_role" "seed_defaults" {
  name = "${var.prefix}-seed-defaults"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })

  tags = local.tags
}

resource "aws_iam_role_policy_attachment" "seed_defaults_basic" {
  role       = aws_iam_role.seed_defaults.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "seed_defaults_dynamo" {
  name = "${var.prefix}-seed-defaults-dynamo"
  role = aws_iam_role.seed_defaults.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:Query"]
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

# Lambda function using the shared binary (same zip as all other Lambdas)
resource "aws_lambda_function" "seed_defaults" {
  function_name    = "${var.prefix}-seed-defaults"
  role             = aws_iam_role.seed_defaults.arn
  filename         = data.archive_file.shared_lambda_zip.output_path
  source_code_hash = data.archive_file.shared_lambda_zip.output_base64sha256
  handler          = "bootstrap"
  runtime          = "provided.al2"
  timeout          = 60

  environment {
    variables = {
      HANDLER                       = "SeedDefaults"
      ROLE_DEFINITIONS_TABLE_NAME   = aws_dynamodb_table.role_definitions.name
      OWNERSHIP_POLICIES_TABLE_NAME = aws_dynamodb_table.ownership_policies.name
    }
  }

  tags = local.tags

  depends_on = [data.archive_file.shared_lambda_zip]
}

# Invocation — runs after tables exist. Bump seed_version to re-seed.
resource "aws_lambda_invocation" "seed_defaults" {
  function_name = aws_lambda_function.seed_defaults.function_name
  input         = jsonencode({})

  triggers = {
    seed_version = "1"
  }

  depends_on = [
    aws_dynamodb_table.role_definitions,
    aws_dynamodb_table.ownership_policies,
    aws_lambda_function.seed_defaults,
  ]
}
```

- [ ] **Step 2: Validate Terraform**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals && terraform init -backend=false -reconfigure 2>&1 | tail -3 && terraform validate
```

Expected: `Success! The configuration is valid.`

- [ ] **Step 3: Commit**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals
git add seed_defaults.tf
git commit -m "feat: add Terraform seed Lambda invocation for global defaults"
```

---

## Task 5: Tenant management router

**Files:**
- Create: `files/src/router/tenants/tenants_router.go`

Implements 6 handlers: CreateTenant, GetTenant, UpdateTenant, DeleteTenant, AddTenantMember, RemoveTenantMember.

Authorization rules:
- `CreateTenant`, `DeleteTenant`: global `IsAdmin` only
- `GetTenant`: global `IsAdmin` OR tenant member (any role)
- `UpdateTenant`: global `IsAdmin` OR `IsTenantAdmin`
- `AddTenantMember`, `RemoveTenantMember`: global `IsAdmin` OR `IsTenantAdmin`

- [ ] **Step 1: Create `files/src/router/tenants/tenants_router.go`**

Create `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\router\tenants\tenants_router.go`:

```go
package tenants

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

type createTenantRequest struct {
	Name string `json:"name"`
}

type updateTenantRequest struct {
	Name *string `json:"name,omitempty"`
}

type addTenantMemberRequest struct {
	UserId string                  `json:"userId"`
	Role   models.TenantMemberRole `json:"role"`
}

func CreateTenant(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var req createTenantRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || req.Name == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	ownerId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
	tenant, err := db.CreateTenant(ctx, req.Name, ownerId)
	if err != nil {
		log.WithError(err).Error("CreateTenant db error")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{"tenant": tenant})
}

func GetTenant(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		userId := utils.GetCognitoUsername(event.RequestContext.Authorizer)
		member, err := db.GetTenantMemberByUserAndTenant(ctx, userId, tenantId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if member == nil || member.Status != models.TenantMemberStatusActive {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	tenant, err := db.GetTenantById(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if tenant == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantNotFound, nil)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"tenant": tenant})
}

func UpdateTenant(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		isTA, err := db.IsTenantAdmin(ctx, utils.GetCognitoUsername(event.RequestContext.Authorizer), tenantId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if !isTA {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	tenant, err := db.GetTenantById(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if tenant == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantNotFound, nil)
	}
	var req updateTenantRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if err := db.UpdateTenant(ctx, tenant); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"tenant": tenant})
}

func DeleteTenant(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	tenant, err := db.GetTenantById(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if tenant == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantNotFound, nil)
	}
	if err := db.DeleteTenantById(ctx, tenantId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}

func AddTenantMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		isTA, err := db.IsTenantAdmin(ctx, utils.GetCognitoUsername(event.RequestContext.Authorizer), tenantId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if !isTA {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	var req addTenantMemberRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || req.UserId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	if req.Role == "" {
		req.Role = models.TenantMemberRoleMember
	}
	member, err := db.AddTenantMember(ctx, tenantId, req.UserId, req.Role)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{"member": member})
}

func RemoveTenantMember(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	memberId := event.PathParameters["memberId"]
	if tenantId == "" || memberId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	if !utils.IsAdmin(event.RequestContext.Authorizer) {
		isTA, err := db.IsTenantAdmin(ctx, utils.GetCognitoUsername(event.RequestContext.Authorizer), tenantId)
		if err != nil {
			return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
		}
		if !isTA {
			return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
		}
	}
	member, err := db.GetTenantMemberById(ctx, memberId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if member == nil || member.TenantId != tenantId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantMemberNotFound, nil)
	}
	if err := db.RemoveTenantMember(ctx, memberId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}
```

- [ ] **Step 2: Build to verify**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go build ./router/tenants/...
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals
git add files/src/router/tenants/tenants_router.go
git commit -m "feat: add tenant management router handlers"
```

---

## Task 6: Role definitions + ownership policies routers

**Files:**
- Create: `files/src/router/tenants/roles_router.go`
- Create: `files/src/router/tenants/ownership_policies_router.go`

Authorization for all handlers in these files: global `IsAdmin` OR `IsTenantAdmin`.

- [ ] **Step 1: Create `files/src/router/tenants/roles_router.go`**

Create `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\router\tenants\roles_router.go`:

```go
package tenants

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

type createRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type updateRoleRequest struct {
	Permissions []string `json:"permissions"`
}

func isTenantAuthorized(ctx context.Context, event events.APIGatewayProxyRequest, tenantId string) (bool, error) {
	if utils.IsAdmin(event.RequestContext.Authorizer) {
		return true, nil
	}
	return db.IsTenantAdmin(ctx, utils.GetCognitoUsername(event.RequestContext.Authorizer), tenantId)
}

func ListRoleDefinitions(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	roles, err := db.ListRoleDefinitionsByTenant(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"roles": roles})
}

func CreateRoleDefinition(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var req createRoleRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || req.Name == "" || len(req.Permissions) == 0 {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	role, err := db.CreateRoleDefinition(ctx, tenantId, req.Name, req.Permissions, false)
	if err != nil {
		log.WithError(err).Error("CreateRoleDefinition db error")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{"role": role})
}

func UpdateRoleDefinition(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	roleId := event.PathParameters["roleId"]
	if tenantId == "" || roleId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	existing, err := db.GetRoleDefinitionById(ctx, roleId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if existing == nil || existing.TenantId != tenantId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorRoleNotFound, nil)
	}
	if existing.IsDefault {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorRoleIsDefault, nil)
	}
	var req updateRoleRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || len(req.Permissions) == 0 {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	updated, err := db.UpdateRoleDefinitionPermissions(ctx, roleId, req.Permissions)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"role": updated})
}

func DeleteRoleDefinition(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	roleId := event.PathParameters["roleId"]
	if tenantId == "" || roleId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	existing, err := db.GetRoleDefinitionById(ctx, roleId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if existing == nil || existing.TenantId != tenantId {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorRoleNotFound, nil)
	}
	if existing.IsDefault {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorRoleIsDefault, nil)
	}
	if err := db.DeleteRoleDefinition(ctx, roleId); err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, nil)
}
```

- [ ] **Step 2: Create `files/src/router/tenants/ownership_policies_router.go`**

Create `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\router\tenants\ownership_policies_router.go`:

```go
package tenants

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/utils"
)

type updateOwnershipPolicyRequest struct {
	OwnerPermissions       []string `json:"ownerPermissions"`
	ParentOwnerPermissions []string `json:"parentOwnerPermissions"`
}

func ListOwnershipPolicies(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	policies, err := db.ListOwnershipPoliciesByTenant(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"policies": policies})
}

func UpdateOwnershipPolicy(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	resourceType := event.PathParameters["resourceType"]
	if tenantId == "" || resourceType == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	var req updateOwnershipPolicyRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	policy, err := db.UpsertOwnershipPolicy(ctx, tenantId, resourceType, req.OwnerPermissions, req.ParentOwnerPermissions)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusOK, utils.MsgSuccess, map[string]interface{}{"policy": policy})
}
```

- [ ] **Step 3: Build to verify**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go build ./router/tenants/...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals
git add files/src/router/tenants/roles_router.go files/src/router/tenants/ownership_policies_router.go
git commit -m "feat: add role definitions and ownership policies management handlers"
```

---

## Task 7: Tenanted team creation handler

**Files:**
- Create: `files/src/router/tenants/tenanted_teams_router.go`

`POST /v1/tenants/{tenantId}/teams` — creates a team pre-linked to the tenant. Authorization: tenant admin only.

- [ ] **Step 1: Create `files/src/router/tenants/tenanted_teams_router.go`**

Create `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\router\tenants\tenanted_teams_router.go`:

```go
package tenants

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	log "github.com/sirupsen/logrus"

	"github.com/fpgschiba/volleygoals/db"
	"github.com/fpgschiba/volleygoals/models"
	"github.com/fpgschiba/volleygoals/utils"
)

type createTenantedTeamRequest struct {
	Name string `json:"name"`
}

// CreateTenantedTeam creates a new team already linked to the given tenant.
// Authorization: global ADMINS or tenant admin.
func CreateTenantedTeam(ctx context.Context, event events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	tenantId := event.PathParameters["tenantId"]
	if tenantId == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, nil)
	}
	ok, err := isTenantAuthorized(ctx, event, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if !ok {
		return utils.ErrorResponse(http.StatusForbidden, utils.MsgErrorForbidden, nil)
	}
	tenant, err := db.GetTenantById(ctx, tenantId)
	if err != nil {
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	if tenant == nil {
		return utils.ErrorResponse(http.StatusNotFound, utils.MsgErrorTenantNotFound, nil)
	}
	var req createTenantedTeamRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil || req.Name == "" {
		return utils.ErrorResponse(http.StatusBadRequest, utils.MsgBadRequest, err)
	}
	team, err := db.CreateTeam(ctx, req.Name, tenantId)
	if err != nil {
		log.WithError(err).Error("CreateTenantedTeam db error")
		return utils.ErrorResponse(http.StatusInternalServerError, utils.MsgInternalServerError, err)
	}
	return utils.SuccessResponse(http.StatusCreated, utils.MsgSuccess, map[string]interface{}{"team": team})
}
```

> **Note:** `db.CreateTeam` does not yet accept a `tenantId`. Check `db/teams.go` — if the existing `CreateTeam` function doesn't accept tenantId, add a new `CreateTeamWithTenant(ctx, name, tenantId string) (*models.Team, error)` function following the same pattern as `CreateTeam`, setting `TenantId: &tenantId` on the `Team` struct. Use that instead and adjust the import here to match.

- [ ] **Step 2: Check `db.CreateTeam` signature**

```bash
grep -n "func CreateTeam" D:/Projects/terraform-aws-modules/terraform-volleygoals/files/src/db/teams.go
```

If the signature is `CreateTeam(ctx, name string)` (no tenantId), add `CreateTeamWithTenant` to `db/teams.go`:

```go
func CreateTeamWithTenant(ctx context.Context, name, tenantId string) (*models.Team, error) {
	client = GetClient()
	now := time.Now()
	team := &models.Team{
		Id:        models.GenerateID(),
		Name:      name,
		Status:    models.TeamStatusActive,
		TenantId:  &tenantId,
		CreatedAt: now,
		UpdatedAt: now,
	}
	item, err := attributevalue.MarshalMap(team)
	if err != nil {
		return nil, err
	}
	_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &teamsTableName,
		Item:      item,
	})
	return team, err
}
```

Then update `tenanted_teams_router.go` to call `db.CreateTeamWithTenant(ctx, req.Name, tenantId)` instead of `db.CreateTeam`.

- [ ] **Step 3: Build to verify**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go build ./...
```

Expected: no output.

- [ ] **Step 4: Commit**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals
git add files/src/router/tenants/tenanted_teams_router.go files/src/db/teams.go
git commit -m "feat: add tenanted team creation handler"
```

---

## Task 8: Register all new handlers in router.go

**Files:**
- Modify: `files/src/router/router.go`

- [ ] **Step 1: Add imports to `router/router.go`**

Open `D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src\router\router.go`. Add to the import block:

```go
"github.com/fpgschiba/volleygoals/router/tenants"
```

(`seed` was already added in Task 3.)

- [ ] **Step 2: Add all 13 new handler cases to the switch statement**

Add after the existing `SeedDefaults` case:

```go
	// Tenant management handlers
	case "CreateTenant":
		response, err = tenants.CreateTenant(ctx, event)
	case "GetTenant":
		response, err = tenants.GetTenant(ctx, event)
	case "UpdateTenant":
		response, err = tenants.UpdateTenant(ctx, event)
	case "DeleteTenant":
		response, err = tenants.DeleteTenant(ctx, event)
	case "AddTenantMember":
		response, err = tenants.AddTenantMember(ctx, event)
	case "RemoveTenantMember":
		response, err = tenants.RemoveTenantMember(ctx, event)

	// Role definition handlers
	case "ListRoleDefinitions":
		response, err = tenants.ListRoleDefinitions(ctx, event)
	case "CreateRoleDefinition":
		response, err = tenants.CreateRoleDefinition(ctx, event)
	case "UpdateRoleDefinition":
		response, err = tenants.UpdateRoleDefinition(ctx, event)
	case "DeleteRoleDefinition":
		response, err = tenants.DeleteRoleDefinition(ctx, event)

	// Ownership policy handlers
	case "ListOwnershipPolicies":
		response, err = tenants.ListOwnershipPolicies(ctx, event)
	case "UpdateOwnershipPolicy":
		response, err = tenants.UpdateOwnershipPolicy(ctx, event)

	// Tenanted team handlers
	case "CreateTenantedTeam":
		response, err = tenants.CreateTenantedTeam(ctx, event)
```

- [ ] **Step 3: Build full project**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go build ./...
```

Expected: no output.

- [ ] **Step 4: Run all tests**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go test ./...
```

Expected: all existing `TestCheckPermission_*` tests still pass.

- [ ] **Step 5: Commit**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals
git add files/src/router/router.go
git commit -m "feat: register all tenant management handlers in router"
```

---

## Task 9: Terraform — routes_tenants.tf

**Files:**
- Create: `routes_tenants.tf`

Creates API Gateway resources and Lambda modules for all 13 tenant management endpoints. Follows the same pattern as `routes_invites.tf`.

- [ ] **Step 1: Create `routes_tenants.tf`**

Create `D:\Projects\terraform-aws-modules\terraform-volleygoals\routes_tenants.tf`:

```hcl
# ─── API Gateway Resources ─────────────────────────────────────────────────

resource "aws_api_gateway_resource" "tenants" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "tenants"
}

resource "aws_api_gateway_resource" "tenant_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.tenants.id
  path_part   = "{tenantId}"
}

resource "aws_api_gateway_resource" "tenant_members" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.tenant_id.id
  path_part   = "members"
}

resource "aws_api_gateway_resource" "tenant_member_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.tenant_members.id
  path_part   = "{memberId}"
}

resource "aws_api_gateway_resource" "tenant_roles" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.tenant_id.id
  path_part   = "roles"
}

resource "aws_api_gateway_resource" "tenant_role_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.tenant_roles.id
  path_part   = "{roleId}"
}

resource "aws_api_gateway_resource" "tenant_ownership_policies" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.tenant_id.id
  path_part   = "ownership-policies"
}

resource "aws_api_gateway_resource" "tenant_ownership_policy_type" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.tenant_ownership_policies.id
  path_part   = "{resourceType}"
}

resource "aws_api_gateway_resource" "tenant_teams" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.tenant_id.id
  path_part   = "teams"
}

# ─── Shared IAM block (permission tables, used by all tenant handlers) ──────

locals {
  tenant_permission_iam = {
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
  }
}

# ─── POST /v1/tenants ────────────────────────────────────────────────────────

module "create_tenant_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["POST"]
  name_overwrite       = "create-tenant"
  path_name            = "tenants"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenants.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "CreateTenant"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.tenants.arn]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenants, data.archive_file.shared_lambda_zip]
}

# ─── GET /v1/tenants/{tenantId} ─────────────────────────────────────────────

module "get_tenant_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["GET"]
  name_overwrite       = "get-tenant"
  path_name            = "tenants"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_id.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "GetTenant"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.tenants.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_id, data.archive_file.shared_lambda_zip]
}

# ─── PATCH /v1/tenants/{tenantId} ───────────────────────────────────────────

module "update_tenant_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["PATCH"]
  name_overwrite       = "update-tenant"
  path_name            = "tenants"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_id.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "UpdateTenant"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:PutItem"]
      resources = [aws_dynamodb_table.tenants.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_id, data.archive_file.shared_lambda_zip]
}

# ─── DELETE /v1/tenants/{tenantId} ──────────────────────────────────────────

module "delete_tenant_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["DELETE"]
  name_overwrite       = "delete-tenant"
  path_name            = "tenants"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_id.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "DeleteTenant"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.tenants.arn]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_id, data.archive_file.shared_lambda_zip]
}

# ─── POST /v1/tenants/{tenantId}/members ────────────────────────────────────

module "add_tenant_member_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["POST"]
  name_overwrite       = "add-tenant-member"
  path_name            = "members"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_members.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "AddTenantMember"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.tenant_members.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_members, data.archive_file.shared_lambda_zip]
}

# ─── DELETE /v1/tenants/{tenantId}/members/{memberId} ───────────────────────

module "remove_tenant_member_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["DELETE"]
  name_overwrite       = "remove-tenant-member"
  path_name            = "members"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_member_id.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "RemoveTenantMember"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.tenant_members.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_member_id, data.archive_file.shared_lambda_zip]
}

# ─── GET /v1/tenants/{tenantId}/roles ───────────────────────────────────────

module "list_role_definitions_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["GET"]
  name_overwrite       = "list-role-definitions"
  path_name            = "roles"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_roles.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "ListRoleDefinitions"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.role_definitions.arn}/index/tenantIdIndex"]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_roles, data.archive_file.shared_lambda_zip]
}

# ─── POST /v1/tenants/{tenantId}/roles ──────────────────────────────────────

module "create_role_definition_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["POST"]
  name_overwrite       = "create-role-definition"
  path_name            = "roles"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_roles.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "CreateRoleDefinition"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.role_definitions.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_roles, data.archive_file.shared_lambda_zip]
}

# ─── PATCH /v1/tenants/{tenantId}/roles/{roleId} ────────────────────────────

module "update_role_definition_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["PATCH"]
  name_overwrite       = "update-role-definition"
  path_name            = "roles"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_role_id.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "UpdateRoleDefinition"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.role_definitions.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_role_id, data.archive_file.shared_lambda_zip]
}

# ─── DELETE /v1/tenants/{tenantId}/roles/{roleId} ───────────────────────────

module "delete_role_definition_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["DELETE"]
  name_overwrite       = "delete-role-definition"
  path_name            = "roles"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_role_id.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "DeleteRoleDefinition"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.role_definitions.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_role_id, data.archive_file.shared_lambda_zip]
}

# ─── GET /v1/tenants/{tenantId}/ownership-policies ──────────────────────────

module "list_ownership_policies_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["GET"]
  name_overwrite       = "list-ownership-policies"
  path_name            = "ownership-policies"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_ownership_policies.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "ListOwnershipPolicies"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.ownership_policies.arn}/index/tenantIdIndex"]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_ownership_policies, data.archive_file.shared_lambda_zip]
}

# ─── PATCH /v1/tenants/{tenantId}/ownership-policies/{resourceType} ─────────

module "update_ownership_policy_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["PATCH"]
  name_overwrite       = "update-ownership-policy"
  path_name            = "ownership-policies"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_ownership_policy_type.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "UpdateOwnershipPolicy"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem", "dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.ownership_policies.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.ownership_policies.arn}/index/tenantResourceTypeIndex"]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_ownership_policy_type, data.archive_file.shared_lambda_zip]
}

# ─── POST /v1/tenants/{tenantId}/teams ──────────────────────────────────────

module "create_tenanted_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id               = aws_api_gateway_rest_api.api.id
  code_dir             = "${path.module}/files/src"
  cors_enabled         = true
  control_allow_origin = local.cors_allowed_origin
  http_methods         = ["POST"]
  name_overwrite       = "create-tenanted-team"
  path_name            = "teams"
  create_resource      = false
  existing_resource_id = aws_api_gateway_resource.tenant_teams.id
  prefix               = var.prefix
  authorizer_id        = aws_api_gateway_authorizer.this.id
  authorization_type   = "COGNITO_USER_POOLS"
  enable_tracing       = true
  timeout              = 29
  vpc_networked        = false
  environment_variables = local.lambda_environment_variables
  tags                 = local.tags
  layer_arns           = local.lambda_layer_arns
  json_logging         = true
  handler_name         = "CreateTenantedTeam"
  pre_built_zip        = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.teams.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.tenants.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.tenant_members.arn}/index/tenantUserIndex"]
    },
    local.tenant_permission_iam,
  ]

  depends_on = [aws_api_gateway_rest_api.api, aws_api_gateway_resource.tenant_teams, data.archive_file.shared_lambda_zip]
}
```

- [ ] **Step 2: Validate Terraform**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals && terraform init -backend=false -reconfigure 2>&1 | tail -3 && terraform validate
```

Expected: `Success! The configuration is valid.`

- [ ] **Step 3: Commit**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals
git add routes_tenants.tf
git commit -m "feat: add Terraform routes for all tenant management endpoints"
```

---

## Task 10: Final build, tests, and branch clean-up

**Files:** None (verification only)

- [ ] **Step 1: Full Go build**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go build ./...
```

Expected: no output.

- [ ] **Step 2: Run all tests**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals\files\src && go test ./... -v
```

Expected: all `TestCheckPermission_*` tests PASS. No failures.

- [ ] **Step 3: Final Terraform validation**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals && terraform validate
```

Expected: `Success! The configuration is valid.`

- [ ] **Step 4: Git log review**

```bash
cd D:\Projects\terraform-aws-modules\terraform-volleygoals && git log --oneline feature/dynamic-permissions-model ^main
```

Verify all Plan B commits are present after the Plan A commits.

---

## Done — Plan B Complete

When all tasks are complete:
- Global default permissions are automatically seeded by Terraform via `aws_lambda_invocation` (bump `seed_version = "2"` to re-seed)
- Tenant admins can manage their own role definitions and ownership policies via the new API
- All 13 new endpoints are live in API Gateway with proper IAM and Cognito authorization
- The `feature/dynamic-permissions-model` branch is ready for PR to main
