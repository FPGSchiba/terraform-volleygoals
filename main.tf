module "mc_test" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.2.3"

  api_name           = aws_api_gateway_rest_api.api.name
  code_dir           = "${path.module}/files/src"
  go_build_tags      = ["connection"]
  cors_enabled       = true
  http_methods       = ["GET"]
  path_name          = "test"
  prefix             = var.prefix
  authorizer_id      = aws_api_gateway_authorizer.this.id
  authorization_type = "COGNITO_USER_POOLS"
  enable_tracing     = true
  timeout            = 29
  vpc_networked      = false

  environment_variables = {
    "TABLE_TEST_NAME" = aws_dynamodb_table.test.name
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:DescribeTable",
      ]
      resources = [
        aws_dynamodb_table.test.arn
      ]
    }
  ]

  tags = merge(
    {
      "Application" = "volleygoals"
    },
    var.tags,
  )

  depends_on = [
    aws_api_gateway_rest_api.api
  ]
}
