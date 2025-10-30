module "get_teams_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.2.3"

  api_name              = aws_api_gateway_rest_api.api.name
  code_dir              = "${path.module}/files/src"
  go_build_tags         = ["getTeams"]
  cors_enabled          = true
  http_methods          = ["GET"]
  path_name             = "teams"
  parent_id             = aws_api_gateway_resource.v1.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables

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
    aws_api_gateway_rest_api.api
  ]
}

module "get_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.2.3"

  api_name              = aws_api_gateway_rest_api.api.name
  code_dir              = "${path.module}/files/src"
  go_build_tags         = ["getTeam"]
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "get-team"
  path_name             = "{teamId}"
  parent_id             = module.get_teams_ms.api_resource_id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables

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
    aws_api_gateway_rest_api.api
  ]
}
