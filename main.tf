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
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.0"

  api_name               = aws_api_gateway_rest_api.api.name
  code_dir               = "${path.module}/files/src"
  go_build_tags          = ["getTeams"]
  cors_enabled           = true
  http_methods           = ["GET"]
  name_overwrite         = "list-teams"
  path_name              = "teams"
  existing_resource_path = "/api/v1/teams"
  parent_id              = aws_api_gateway_resource.v1.id
  prefix                 = var.prefix
  authorizer_id          = aws_api_gateway_authorizer.this.id
  authorization_type     = "COGNITO_USER_POOLS"
  enable_tracing         = true
  timeout                = 29
  vpc_networked          = false
  environment_variables  = local.lambda_environment_variables

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:DescribeTable",
      ]
      resources = [
        aws_dynamodb_table.teams.arn
      ]
    }
  ]

  tags = local.tags

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams
  ]
}

module "get_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.0"

  api_name               = aws_api_gateway_rest_api.api.name
  code_dir               = "${path.module}/files/src"
  go_build_tags          = ["getTeam"]
  cors_enabled           = true
  http_methods           = ["GET"]
  name_overwrite         = "get-team"
  path_name              = "{teamId}"
  existing_resource_path = "/api/v1/teams/{teamId}"
  parent_id              = module.get_teams_ms.api_resource_id
  prefix                 = var.prefix
  authorizer_id          = aws_api_gateway_authorizer.this.id
  authorization_type     = "COGNITO_USER_POOLS"
  enable_tracing         = true
  timeout                = 29
  vpc_networked          = false
  environment_variables  = local.lambda_environment_variables

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:DescribeTable",
      ]
      resources = [
        aws_dynamodb_table.teams.arn
      ]
    }
  ]

  tags = local.tags

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams_id
  ]
}

module "create_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.0"

  api_name               = aws_api_gateway_rest_api.api.name
  code_dir               = "${path.module}/files/src"
  go_build_tags          = ["createTeam"]
  cors_enabled           = false # already handled in get_teams_ms
  http_methods           = ["POST"]
  name_overwrite         = "create-team"
  path_name              = "teams"
  existing_resource_path = "/api/v1/teams"
  parent_id              = aws_api_gateway_resource.v1.id
  prefix                 = var.prefix
  authorizer_id          = aws_api_gateway_authorizer.this.id
  authorization_type     = "COGNITO_USER_POOLS"
  enable_tracing         = true
  timeout                = 29
  vpc_networked          = false
  environment_variables  = local.lambda_environment_variables

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:PutItem",
      ]
      resources = [
        aws_dynamodb_table.teams.arn,
        "${aws_dynamodb_table.team_members.arn}/*",
      ]
    }
  ]

  tags = local.tags

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams,
    module.get_teams_ms
  ]
}

module "delete_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.0"

  api_name               = aws_api_gateway_rest_api.api.name
  code_dir               = "${path.module}/files/src"
  go_build_tags          = ["deleteTeam"]
  cors_enabled           = false # already handled in get_team_ms
  http_methods           = ["DELETE"]
  name_overwrite         = "delete-team"
  path_name              = "{teamId}"
  existing_resource_path = "/api/v1/teams/{teamId}"
  parent_id              = aws_api_gateway_resource.teams.id
  prefix                 = var.prefix
  authorizer_id          = aws_api_gateway_authorizer.this.id
  authorization_type     = "COGNITO_USER_POOLS"
  enable_tracing         = true
  timeout                = 29
  vpc_networked          = false
  environment_variables  = local.lambda_environment_variables

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:PutItem",
      ]
      resources = [
        aws_dynamodb_table.teams.arn,
        "${aws_dynamodb_table.team_members.arn}/*",
      ]
    }
  ]

  tags = local.tags

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams_id,
    module.get_team_ms
  ]
}

module "update_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.0"

  api_name               = aws_api_gateway_rest_api.api.name
  code_dir               = "${path.module}/files/src"
  go_build_tags          = ["updateTeam"]
  cors_enabled           = false # already handled in get_team_ms
  http_methods           = ["PATCH"]
  name_overwrite         = "update-team"
  path_name              = "{teamId}"
  existing_resource_path = "/api/v1/teams/{teamId}"
  parent_id              = aws_api_gateway_resource.teams.id
  prefix                 = var.prefix
  authorizer_id          = aws_api_gateway_authorizer.this.id
  authorization_type     = "COGNITO_USER_POOLS"
  enable_tracing         = true
  timeout                = 29
  vpc_networked          = false
  environment_variables  = local.lambda_environment_variables

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:PutItem",
      ]
      resources = [
        aws_dynamodb_table.teams.arn,
        "${aws_dynamodb_table.team_members.arn}/*",
      ]
    }
  ]

  tags = local.tags

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams_id,
    module.get_team_ms
  ]
}

# Team Members
