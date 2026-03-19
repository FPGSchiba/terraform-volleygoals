# Self
resource "aws_api_gateway_resource" "self" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "self"
}

module "get_self_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "get-self"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.self.id
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

  handler_name  = "GetSelf"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

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
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.teams.arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.self,
    data.archive_file.shared_lambda_zip,
  ]
}

module "update_self_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["PATCH"]
  name_overwrite        = "update-self"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.self.id
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

  handler_name  = "UpdateSelf"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["cognito-idp:AdminUpdateUserAttributes", "cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.self,
    data.archive_file.shared_lambda_zip,
  ]
}

# Teams
resource "aws_api_gateway_resource" "teams" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "teams"
}

resource "aws_api_gateway_resource" "teams_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams.id
  path_part   = "{teamId}"
}

module "get_teams_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "list-teams"
  path_name             = "teams"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams.id
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

  handler_name  = "ListTeams"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Scan",
      ]
      resources = [
        aws_dynamodb_table.teams.arn
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams,
    data.archive_file.shared_lambda_zip,
  ]
}

module "get_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "get-team"
  path_name             = "{teamId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_id.id
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

  handler_name  = "GetTeam"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:GetItem"]
      resources = [aws_dynamodb_table.teams.arn]
    },
    {
      actions = ["dynamodb:Query"]
      resources = [
        "${aws_dynamodb_table.team_settings.arn}/index/teamIdIndex",
        "${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex",
      ]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams_id,
    data.archive_file.shared_lambda_zip,
  ]
}

module "create_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["POST"]
  name_overwrite        = "create-team"
  path_name             = "teams"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = merge(local.lambda_environment_variables, { "LOG_LEVEL" = "debug" })
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true

  handler_name  = "CreateTeam"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem", "dynamodb:Scan"]
      resources = [aws_dynamodb_table.teams.arn]
    },
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.team_settings.arn, aws_dynamodb_table.team_members.arn]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams,
    module.get_teams_ms,
    data.archive_file.shared_lambda_zip,
  ]
}

module "delete_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-team"
  path_name             = "{teamId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_id.id
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

  handler_name  = "DeleteTeam"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["dynamodb:DeleteItem", "dynamodb:GetItem"]
      resources = [aws_dynamodb_table.teams.arn]
    },
    {
      actions = ["dynamodb:Scan", "dynamodb:DeleteItem"]
      resources = [
        aws_dynamodb_table.seasons.arn,
        aws_dynamodb_table.goals.arn,
        aws_dynamodb_table.progress_reports.arn,
        aws_dynamodb_table.progress.arn,
        aws_dynamodb_table.comments.arn,
        aws_dynamodb_table.comment_files.arn,
      ]
    },
    {
      actions = ["dynamodb:Query", "dynamodb:DeleteItem"]
      resources = [
        aws_dynamodb_table.team_members.arn,
        "${aws_dynamodb_table.team_members.arn}/index/teamIdIndex",
        aws_dynamodb_table.invites.arn,
        "${aws_dynamodb_table.invites.arn}/index/teamIdIndex",
        aws_dynamodb_table.team_settings.arn,
        "${aws_dynamodb_table.team_settings.arn}/index/teamIdIndex",
      ]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams_id,
    module.get_team_ms,
    data.archive_file.shared_lambda_zip,
  ]
}

module "update_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["PATCH"]
  name_overwrite        = "update-team"
  path_name             = "{teamId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_id.id
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

  handler_name  = "UpdateTeam"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:PutItem", "dynamodb:GetItem"]
      resources = [aws_dynamodb_table.teams.arn]
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
    aws_api_gateway_resource.teams_id,
    module.get_team_ms,
    data.archive_file.shared_lambda_zip,
  ]
}

# Team Invites (list invites for a specific team)

resource "aws_api_gateway_resource" "team_invites" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_id.id
  path_part   = "invites"
}

module "get_team_invites_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "get-team-invites"
  path_name             = "invites"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_invites.id
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
  handler_name          = "GetTeamInvites"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions = ["dynamodb:Query"]
      resources = [
        "${aws_dynamodb_table.invites.arn}/index/teamIdIndex",
        "${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex",
      ]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_invites,
    data.archive_file.shared_lambda_zip,
  ]
}

# Team Picture presign

resource "aws_api_gateway_resource" "team_picture" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_id.id
  path_part   = "picture"
}

resource "aws_api_gateway_resource" "team_picture_presign" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.team_picture.id
  path_part   = "presign"
}

module "upload_team_picture_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "upload-team-picture"
  path_name             = "presign"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_picture_presign.id
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
  handler_name          = "UploadTeamPicture"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["s3:PutObject"]
      resources = ["${aws_s3_bucket.this.arn}/teams/*"]
    },
    {
      actions   = ["dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.teams.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_picture_presign,
    data.archive_file.shared_lambda_zip,
  ]
}

# Team Activity

resource "aws_api_gateway_resource" "team_activity" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_id.id
  path_part   = "activity"
}

module "get_team_activity_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "get-team-activity"
  path_name             = "activity"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_activity.id
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
  handler_name          = "GetTeamActivity"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Scan"]
      resources = [aws_dynamodb_table.activities.arn]
    },
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_activity,
    data.archive_file.shared_lambda_zip,
  ]
}

# Team Settings

resource "aws_api_gateway_resource" "team_settings" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_id.id
  path_part   = "settings"
}

module "update_team_settings_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["PATCH"]
  name_overwrite        = "update-team-settings"
  path_name             = "settings"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_settings.id
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

  handler_name  = "UpdateTeamSettings"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions = ["dynamodb:Query", "dynamodb:PutItem"]
      resources = [
        aws_dynamodb_table.team_settings.arn,
        "${aws_dynamodb_table.team_settings.arn}/index/teamIdIndex",
      ]
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
    aws_api_gateway_resource.team_settings,
    data.archive_file.shared_lambda_zip,
  ]
}

# Team Members

resource "aws_api_gateway_resource" "team_members" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_id.id
  path_part   = "members"
}

resource "aws_api_gateway_resource" "team_member_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.team_members.id
  path_part   = "{memberId}"
}

module "list_team_members_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "list-team-members"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_members.id
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

  handler_name  = "ListTeamMembers"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions = ["dynamodb:Scan", "dynamodb:Query"]
      resources = [
        aws_dynamodb_table.team_members.arn,
        "${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex",
      ]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_members,
    data.archive_file.shared_lambda_zip,
  ]
}

module "add_team_member_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["POST"]
  name_overwrite        = "add-team-member"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_members.id
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

  handler_name  = "AddTeamMember"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.team_members.arn]
    },
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.activities.arn]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_members,
    data.archive_file.shared_lambda_zip,
  ]
}

module "update_team_member_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["PATCH"]
  name_overwrite        = "update-team-member"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_member_id.id
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

  handler_name  = "UpdateTeamMember"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["dynamodb:UpdateItem"]
      resources = [aws_dynamodb_table.team_members.arn]
    },
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.activities.arn]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_settings,
    data.archive_file.shared_lambda_zip,
  ]
}

module "delete_team_member_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-team-member"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_member_id.id
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

  handler_name  = "RemoveTeamMember"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["dynamodb:Query"]
      resources = ["${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex"]
    },
    {
      actions   = ["dynamodb:DeleteItem"]
      resources = [aws_dynamodb_table.team_members.arn]
    },
    {
      actions   = ["dynamodb:PutItem"]
      resources = [aws_dynamodb_table.activities.arn]
    },
    {
      actions   = ["cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_settings,
    data.archive_file.shared_lambda_zip,
  ]
}

module "leave_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["DELETE"]
  name_overwrite        = "leave-team"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_members.id
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

  handler_name  = "LeaveTeam"
  pre_built_zip = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions = ["dynamodb:Query", "dynamodb:UpdateItem"]
      resources = [
        aws_dynamodb_table.team_members.arn,
        "${aws_dynamodb_table.team_members.arn}/index/teamUserIdIndex",
        "${aws_dynamodb_table.team_members.arn}/index/teamIdIndex",
      ]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_settings,
    data.archive_file.shared_lambda_zip,
  ]
}
