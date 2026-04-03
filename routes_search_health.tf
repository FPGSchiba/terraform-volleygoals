# Search

resource "aws_api_gateway_resource" "search" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "search"
}

module "global_search_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.2"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "global-search"
  path_name             = "search"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.search.id
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
  handler_name          = "GlobalSearch"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
    {
      actions = ["dynamodb:Scan", "dynamodb:Query"]
      resources = [
        aws_dynamodb_table.teams.arn,
        aws_dynamodb_table.team_members.arn,
        aws_dynamodb_table.seasons.arn,
        aws_dynamodb_table.goals.arn,
        aws_dynamodb_table.comments.arn,
        aws_dynamodb_table.progress_reports.arn,
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
      ]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.search,
    data.archive_file.shared_lambda_zip,
  ]
}

# Health check (public, no auth)

resource "aws_api_gateway_resource" "health" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "health"
}

module "health_check_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.2"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["POST"]
  name_overwrite        = "health-check"
  path_name             = "health"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.health.id
  prefix                = var.prefix
  authorization_type    = "NONE"
  enable_tracing        = false
  timeout               = 10
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "HealthCheck"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path
  runtime               = local.lambda_runtime

  additional_iam_statements = [
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
    aws_api_gateway_resource.health,
    data.archive_file.shared_lambda_zip,
  ]
}
