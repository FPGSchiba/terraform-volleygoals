# ─── API Gateway Resources ────────────────────────────────────────────────────

resource "aws_api_gateway_resource" "teams_teamId_goals" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId.id
  path_part   = "goals"
}

resource "aws_api_gateway_resource" "teams_teamId_goals_goalId" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId_goals.id
  path_part   = "{goalId}"
}

resource "aws_api_gateway_resource" "teams_teamId_goals_goalId_seasons" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId_goals_goalId.id
  path_part   = "seasons"
}

resource "aws_api_gateway_resource" "teams_teamId_goals_goalId_seasons_seasonId" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId_goals_goalId_seasons.id
  path_part   = "{seasonId}"
}

resource "aws_api_gateway_resource" "teams_teamId_goals_goalId_picture" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_teamId_goals_goalId.id
  path_part   = "picture"
}

# ─── Goal modules ─────────────────────────────────────────────────────────────

module "create_goal_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["POST"]
  name_overwrite        = "create-goal"
  path_name             = "goals"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals.id
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
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:PutItem"]
      Resource = [aws_dynamodb_table.goals.arn]
    }
  ]
}

module "list_goals_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["GET"]
  name_overwrite        = "list-goals"
  path_name             = "goals"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals.id
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
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:Scan", "dynamodb:Query"]
      Resource = [
        aws_dynamodb_table.goals.arn,
        aws_dynamodb_table.goal_seasons.arn,
        "${aws_dynamodb_table.goal_seasons.arn}/index/seasonIdIndex",
        aws_dynamodb_table.progress.arn,
      ]
    }
  ]
}

module "get_goal_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["GET"]
  name_overwrite        = "get-goal"
  path_name             = "{goalId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId.id
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
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem"]
      Resource = [aws_dynamodb_table.goals.arn]
    }
  ]
}

module "update_goal_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["PUT"]
  name_overwrite        = "update-goal"
  path_name             = "{goalId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId.id
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
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:UpdateItem"]
      Resource = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.activities.arn]
    }
  ]
}

module "delete_goal_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-goal"
  path_name             = "{goalId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId.id
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
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      Resource = [aws_dynamodb_table.goals.arn]
    }
  ]
}

module "upload_goal_picture_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["POST"]
  name_overwrite        = "upload-goal-picture"
  path_name             = "picture"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId_picture.id
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
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:UpdateItem", "s3:PutObject"]
      Resource = [aws_dynamodb_table.goals.arn, "${aws_s3_bucket.this.arn}/*"]
    }
  ]
}

# ─── Goal Season tagging modules ─────────────────────────────────────────────

module "tag_goal_season_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["POST"]
  name_overwrite        = "tag-goal-season"
  path_name             = "{seasonId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId_seasons_seasonId.id
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
  handler_name          = "TagGoalToSeason"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:PutItem"]
      Resource = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.goal_seasons.arn]
    }
  ]
}

module "untag_goal_season_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["DELETE"]
  name_overwrite        = "untag-goal-season"
  path_name             = "{seasonId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId_seasons_seasonId.id
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
  handler_name          = "UntagGoalFromSeason"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:DeleteItem"]
      Resource = [aws_dynamodb_table.goals.arn, aws_dynamodb_table.goal_seasons.arn]
    }
  ]
}

module "list_goal_seasons_ms" {
  source                = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.1"
  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  create_options_method = false
  http_methods          = ["GET"]
  name_overwrite        = "list-goal-seasons"
  path_name             = "seasons"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_teamId_goals_goalId_seasons.id
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
  handler_name          = "ListGoalSeasons"
  runtime               = local.lambda_runtime
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      Effect   = "Allow"
      Action   = ["dynamodb:GetItem", "dynamodb:Query"]
      Resource = [
        aws_dynamodb_table.goals.arn,
        aws_dynamodb_table.goal_seasons.arn,
        "${aws_dynamodb_table.goal_seasons.arn}/index/goalIdIndex",
      ]
    }
  ]
}
