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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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
  runtime       = local.lambda_runtime

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