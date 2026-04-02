# Users (admin only)

resource "aws_api_gateway_resource" "users" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "users"
}

resource "aws_api_gateway_resource" "user_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.users.id
  path_part   = "{userSub}"
}

module "list_users_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "list-users"
  path_name             = "users"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.users.id
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
  handler_name          = "ListUsers"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime       = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["cognito-idp:ListUsers", "cognito-idp:AdminListGroupsForUser", "cognito-idp:ListUsersInGroup"]
      resources = [var.cognito_user_pool_arn]
    },
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
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.users,
    data.archive_file.shared_lambda_zip,
  ]
}

module "get_user_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "get-user"
  path_name             = "{userSub}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.user_id.id
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
  handler_name          = "GetUser"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime       = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/userIdIndex"]
    },
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
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.user_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "delete_user_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-user"
  path_name             = "{userSub}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.user_id.id
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
  handler_name          = "DeleteUser"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime       = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["cognito-idp:AdminDeleteUser"]
      resources = [var.cognito_user_pool_arn]
    },
    {
      actions   = ["dynamodb:Query", "dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.team_members.arn, "${aws_dynamodb_table.team_members.arn}/index/userIdIndex"]
    },
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
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.user_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "update_user_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["PATCH"]
  name_overwrite        = "update-user"
  path_name             = "{userSub}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.user_id.id
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
  handler_name          = "UpdateUser"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime       = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["cognito-idp:AdminUpdateUserAttributes", "cognito-idp:AdminGetUser", "cognito-idp:AdminAddUserToGroup", "cognito-idp:AdminRemoveUserFromGroup", "cognito-idp:AdminEnableUser", "cognito-idp:AdminDisableUser"]
      resources = [var.cognito_user_pool_arn]
    },
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
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.user_id,
    data.archive_file.shared_lambda_zip,
  ]
}
