# Seasons

resource "aws_api_gateway_resource" "seasons" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "seasons"
}

resource "aws_api_gateway_resource" "season_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.seasons.id
  path_part   = "{seasonId}"
}

resource "aws_api_gateway_resource" "season_stats" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.season_id.id
  path_part   = "stats"
}

# Goals (nested under seasons)

resource "aws_api_gateway_resource" "season_goals" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.season_id.id
  path_part   = "goals"
}

resource "aws_api_gateway_resource" "goal_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.season_goals.id
  path_part   = "{goalId}"
}

resource "aws_api_gateway_resource" "goal_picture" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.goal_id.id
  path_part   = "picture"
}

resource "aws_api_gateway_resource" "goal_picture_presign" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.goal_picture.id
  path_part   = "presign"
}

# Progress Reports (nested under seasons)

resource "aws_api_gateway_resource" "season_progress_reports" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.season_id.id
  path_part   = "progress-reports"
}

resource "aws_api_gateway_resource" "progress_report_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.season_progress_reports.id
  path_part   = "{reportId}"
}

# ─── Season modules ──────────────────────────────────────────────────────────

module "create_season_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["POST"]
  name_overwrite        = "create-season"
  path_name             = "seasons"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.seasons.id
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
  handler_name          = "CreateSeason"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.seasons,
    data.archive_file.shared_lambda_zip,
  ]
}

module "list_seasons_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "list-seasons"
  path_name             = "seasons"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.seasons.id
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
  handler_name          = "ListSeasons"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Scan"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.seasons,
    data.archive_file.shared_lambda_zip,
  ]
}

module "get_season_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "get-season"
  path_name             = "{seasonId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.season_id.id
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
  handler_name          = "GetSeason"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.season_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "update_season_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["PATCH"]
  name_overwrite        = "update-season"
  path_name             = "{seasonId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.season_id.id
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
  handler_name          = "UpdateSeason"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:PutItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.season_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "delete_season_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-season"
  path_name             = "{seasonId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.season_id.id
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
  handler_name          = "DeleteSeason"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions = ["dynamodb:Scan", "dynamodb:DeleteItem"]
      resources = [
        aws_dynamodb_table.goals.arn,
        aws_dynamodb_table.progress_reports.arn,
        aws_dynamodb_table.progress.arn,
        aws_dynamodb_table.comments.arn,
        aws_dynamodb_table.comment_files.arn,
      ]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.season_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "get_season_stats_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "get-season-stats"
  path_name             = "stats"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.season_stats.id
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
  handler_name          = "GetSeasonStats"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:Scan", "dynamodb:Query"]
      resources = [aws_dynamodb_table.seasons.arn, aws_dynamodb_table.goals.arn, aws_dynamodb_table.progress_reports.arn, aws_dynamodb_table.progress.arn]
    },
    {
      actions = ["dynamodb:Query"]
      resources = [
        aws_dynamodb_table.team_members.arn,
        "${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex",
      ]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.season_stats,
    data.archive_file.shared_lambda_zip,
  ]
}

# ─── Goal modules ─────────────────────────────────────────────────────────────

module "create_goal_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["POST"]
  name_overwrite        = "create-goal"
  path_name             = "goals"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.season_goals.id
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
  handler_name          = "CreateGoal"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.goals.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.activities.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.season_goals,
    data.archive_file.shared_lambda_zip,
  ]
}

module "list_goals_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "list-goals"
  path_name             = "goals"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.season_goals.id
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
  handler_name          = "ListGoals"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Scan", "dynamodb:Query"]
      resources = [aws_dynamodb_table.goals.arn]
    },
    {
      actions   = ["dynamodb:Scan"]
      resources = [aws_dynamodb_table.progress.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.season_goals,
    data.archive_file.shared_lambda_zip,
  ]
}

module "get_goal_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "get-goal"
  path_name             = "{goalId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.goal_id.id
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
  handler_name          = "GetGoal"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.goal_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "update_goal_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["PATCH"]
  name_overwrite        = "update-goal"
  path_name             = "{goalId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.goal_id.id
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
  handler_name          = "UpdateGoal"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:PutItem"]
      resources = [aws_dynamodb_table.goals.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.activities.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.goal_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "delete_goal_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-goal"
  path_name             = "{goalId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.goal_id.id
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
  handler_name          = "DeleteGoal"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.goals.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.goal_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "upload_goal_file_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "upload-goal-file"
  path_name             = "presign"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.goal_picture_presign.id
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
  handler_name          = "UploadGoalFile"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["s3:PutObject"]
      resources = ["${aws_s3_bucket.this.arn}/goals/*"]
    },
    {
      actions   = ["dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.goals.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.goal_picture_presign,
    data.archive_file.shared_lambda_zip,
  ]
}

# ─── Progress Report modules ──────────────────────────────────────────────────

module "create_progress_report_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["POST"]
  name_overwrite        = "create-progress-report"
  path_name             = "progress-reports"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.season_progress_reports.id
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
  handler_name          = "CreateProgressReport"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.progress_reports.arn, aws_dynamodb_table.progress.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.activities.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.season_progress_reports,
    data.archive_file.shared_lambda_zip,
  ]
}

module "list_progress_reports_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "list-progress-reports"
  path_name             = "progress-reports"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.season_progress_reports.id
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
  handler_name          = "ListProgressReports"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Scan", "dynamodb:Query"]
      resources = [aws_dynamodb_table.progress_reports.arn]
    },
    {
      actions   = ["dynamodb:Scan"]
      resources = [aws_dynamodb_table.progress.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.season_progress_reports,
    data.archive_file.shared_lambda_zip,
  ]
}

module "get_progress_report_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "get-progress-report"
  path_name             = "{reportId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.progress_report_id.id
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
  handler_name          = "GetProgressReport"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.progress_reports.arn, aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Scan"]
      resources = [aws_dynamodb_table.progress.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.progress_report_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "update_progress_report_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["PATCH"]
  name_overwrite        = "update-progress-report"
  path_name             = "{reportId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.progress_report_id.id
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
  handler_name          = "UpdateProgressReport"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.progress_reports.arn]
    },
    {
      actions   = ["dynamodb:DeleteItem", "dynamodb:PutItem"]
      resources = [aws_dynamodb_table.progress.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.progress_report_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "delete_progress_report_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-progress-report"
  path_name             = "{reportId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.progress_report_id.id
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
  handler_name          = "DeleteProgressReport"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.progress_reports.arn]
    },
    {
      actions   = ["dynamodb:Scan", "dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.progress.arn]
    },
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.seasons.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.progress_report_id,
    data.archive_file.shared_lambda_zip,
  ]
}
