module "mc_test" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.2.2"

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
  vpc_id             = module.vpc.vpc_id
  vpc_networked      = true

  environment_variables = {
    "DB_SECRET_ARN" = module.db.secret_arn
  }

  additional_iam_statements = [
    {
      actions = [
        "secretsmanager:GetSecretValue"
      ]
      resources = [
        module.db.secret_arn
      ]
    }
  ]

  security_groups = [
    {
      name        = "${var.prefix}-volleygoals-lambda"
      description = "Security group for VolleyGoals Lambda functions"
      rules = [
        {
          type             = "egress"
          ip_protocol      = "-1"
          ipv4_cidr_blocks = ["0.0.0.0/0"]
          ipv6_cidr_blocks = ["::/0"]
        },
        {
          type             = "ingress"
          from_port        = 5432
          to_port          = 5432
          ip_protocol      = "tcp"
          ipv4_cidr_blocks = ["172.16.0.0/16"] # VPC
          ipv6_cidr_blocks = []                # VPC
        }
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
