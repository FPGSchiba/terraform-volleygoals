# IAM role for seed Lambda
resource "aws_iam_role" "seed_defaults" {
  name = "${var.prefix}-seed-defaults"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })

  tags = local.tags
}

resource "aws_iam_role_policy_attachment" "seed_defaults_basic" {
  role       = aws_iam_role.seed_defaults.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "seed_defaults_dynamo" {
  name = "${var.prefix}-seed-defaults-dynamo"
  role = aws_iam_role.seed_defaults.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:Query"]
        Resource = [
          aws_dynamodb_table.role_definitions.arn,
          "${aws_dynamodb_table.role_definitions.arn}/index/tenantIdIndex",
          "${aws_dynamodb_table.role_definitions.arn}/index/tenantNameIndex",
          aws_dynamodb_table.ownership_policies.arn,
          "${aws_dynamodb_table.ownership_policies.arn}/index/tenantIdIndex",
          "${aws_dynamodb_table.ownership_policies.arn}/index/tenantResourceTypeIndex",
        ]
      }
    ]
  })
}

# Lambda function using the shared binary (same zip as all other Lambdas)
resource "aws_lambda_function" "seed_defaults" {
  function_name    = "${var.prefix}-seed-defaults"
  role             = aws_iam_role.seed_defaults.arn
  filename         = data.archive_file.shared_lambda_zip.output_path
  source_code_hash = data.archive_file.shared_lambda_zip.output_base64sha256
  handler          = "bootstrap"
  runtime          = "provided.al2"
  timeout          = 60

  environment {
    variables = {
      HANDLER                       = "SeedDefaults"
      ROLE_DEFINITIONS_TABLE_NAME   = aws_dynamodb_table.role_definitions.name
      OWNERSHIP_POLICIES_TABLE_NAME = aws_dynamodb_table.ownership_policies.name
    }
  }

  tags = local.tags

  depends_on = [data.archive_file.shared_lambda_zip]
}

# Invocation — runs after tables exist. Bump seed_version to re-seed.
resource "aws_lambda_invocation" "seed_defaults" {
  function_name = aws_lambda_function.seed_defaults.function_name
  input         = jsonencode({})

  triggers = {
    seed_version = "1"
  }

  depends_on = [
    aws_dynamodb_table.role_definitions,
    aws_dynamodb_table.ownership_policies,
    aws_lambda_function.seed_defaults,
  ]
}
