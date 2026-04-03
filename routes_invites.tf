# Invites

resource "aws_api_gateway_resource" "invites" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "invites"
}

# /invites/complete must be defined before /{inviteId} to avoid path conflict
resource "aws_api_gateway_resource" "invite_complete" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.invites.id
  path_part   = "complete"
}

resource "aws_api_gateway_resource" "invite_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.invites.id
  path_part   = "{inviteId}"
}

module "create_invite_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.2"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["POST"]
  name_overwrite        = "create-invite"
  path_name             = "invites"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.invites.id
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
  handler_name          = "CreateInvite"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.invites.arn]
    },
    {
      actions = ["dynamodb:Query"]
      resources = [
        "${aws_dynamodb_table.invites.arn}/index/tokenIndex",
        "${aws_dynamodb_table.team_members.arn}/index/teamIdIndex",
        "${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex",
      ]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.teams.arn]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser", "cognito-idp:ListUsers"]
      resources = [var.cognito_user_pool_arn]
    },
    {
      actions   = ["ses:SendEmail", "ses:SendTemplatedEmail"]
      resources = ["*"]
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
      ]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.invites,
    data.archive_file.shared_lambda_zip,
  ]
}

# POST /invites/complete — no Cognito auth (public endpoint)
module "complete_invite_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.2"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["POST"]
  name_overwrite        = "complete-invite"
  path_name             = "complete"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.invite_complete.id
  prefix                = var.prefix
  authorization_type    = "NONE"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "CompleteInvite"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Query", "dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.invites.arn, "${aws_dynamodb_table.invites.arn}/index/tokenIndex"]
    },
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.team_members.arn, aws_dynamodb_table.activities.arn]
    },
    {
      actions   = ["cognito-idp:AdminCreateUser", "cognito-idp:AdminAddUserToGroup", "cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser", "cognito-idp:AdminDeleteUser", "cognito-idp:ListUsers"]
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
    aws_api_gateway_resource.invite_complete,
    data.archive_file.shared_lambda_zip,
  ]
}

module "revoke_invite_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.2"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["DELETE"]
  name_overwrite        = "revoke-invite"
  path_name             = "{inviteId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.invite_id.id
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
  handler_name          = "RevokeInvite"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:UpdateItem", "dynamodb:GetItem"]
      resources = [aws_dynamodb_table.invites.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
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
    aws_api_gateway_resource.invite_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "resend_invite_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.2"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["PATCH"]
  name_overwrite        = "resend-invite"
  path_name             = "{inviteId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.invite_id.id
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
  handler_name          = "ResendInvite"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.invites.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.teams.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
    {
      actions   = ["ses:SendEmail", "ses:SendTemplatedEmail"]
      resources = ["*"]
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
      ]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.invite_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "get_invite_by_token_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.2"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["GET"]
  name_overwrite        = "get-invite-by-token"
  path_name             = "{inviteId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.invite_id.id
  prefix                = var.prefix
  authorization_type    = "NONE"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "GetInviteByToken"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions = ["dynamodb:Query"]
      resources = [
        "${aws_dynamodb_table.invites.arn}/index/tokenIndex",
        "${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex",
      ]
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
    aws_api_gateway_resource.invite_id,
    data.archive_file.shared_lambda_zip,
  ]
}
