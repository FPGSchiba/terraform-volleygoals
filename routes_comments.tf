# Comments

resource "aws_api_gateway_resource" "comments" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "comments"
}

resource "aws_api_gateway_resource" "comment_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.comments.id
  path_part   = "{commentId}"
}

resource "aws_api_gateway_resource" "comment_file" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.comment_id.id
  path_part   = "file"
}

resource "aws_api_gateway_resource" "comment_file_presign" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.comment_file.id
  path_part   = "presign"
}

module "create_comment_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["POST"]
  name_overwrite        = "create-comment"
  path_name             = "comments"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.comments.id
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
  handler_name          = "CreateComment"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.comments.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.progress_reports.arn, aws_dynamodb_table.progress.arn, aws_dynamodb_table.seasons.arn]
    },
    {
      actions = ["dynamodb:Query"]
      resources = [
        "${aws_dynamodb_table.team_settings.arn}/index/teamIdIndex",
        "${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex",
      ]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
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
    aws_api_gateway_resource.comments,
    data.archive_file.shared_lambda_zip,
  ]
}

module "list_comments_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["GET"]
  name_overwrite        = "list-comments"
  path_name             = "comments"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.comments.id
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
  handler_name          = "ListComments"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Scan", "dynamodb:Query"]
      resources = [aws_dynamodb_table.comments.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.progress_reports.arn, aws_dynamodb_table.progress.arn, aws_dynamodb_table.seasons.arn]
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
    aws_api_gateway_resource.comments,
    data.archive_file.shared_lambda_zip,
  ]
}

module "get_comment_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "get-comment"
  path_name             = "{commentId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.comment_id.id
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
  handler_name          = "GetComment"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.comments.arn, aws_dynamodb_table.goals.arn, aws_dynamodb_table.progress_reports.arn, aws_dynamodb_table.progress.arn, aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Scan"]
      resources = [aws_dynamodb_table.comment_files.arn]
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
    aws_api_gateway_resource.comment_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "update_comment_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["PATCH"]
  name_overwrite        = "update-comment"
  path_name             = "{commentId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.comment_id.id
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
  handler_name          = "UpdateComment"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.comments.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.progress_reports.arn, aws_dynamodb_table.progress.arn, aws_dynamodb_table.seasons.arn]
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
    aws_api_gateway_resource.comment_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "delete_comment_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-comment"
  path_name             = "{commentId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.comment_id.id
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
  handler_name          = "DeleteComment"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.comments.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.progress_reports.arn, aws_dynamodb_table.progress.arn, aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Scan", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.comment_files.arn]
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
    aws_api_gateway_resource.comment_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "upload_comment_file_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "upload-comment-file"
  path_name             = "presign"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.comment_file_presign.id
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
  handler_name          = "UploadCommentFile"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.comment_files.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.comments.arn, aws_dynamodb_table.goals.arn, aws_dynamodb_table.progress_reports.arn, aws_dynamodb_table.progress.arn, aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["s3:PutObject"]
      resources = ["${aws_s3_bucket.this.arn}/comments/*"]
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
    aws_api_gateway_resource.comment_file_presign,
    data.archive_file.shared_lambda_zip,
  ]
}
