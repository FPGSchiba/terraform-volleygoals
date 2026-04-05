resource "aws_iam_role" "migrate_goals" {
  name = "${var.prefix}-migrate-goals"
  assume_role_policy = jsonencode({
    Version = "2012-10-17", Statement = [{Action = "sts:AssumeRole", Effect = "Allow", Principal = { Service = "lambda.amazonaws.com" }}]
  })
}

resource "aws_iam_role_policy_attachment" "migrate_goals_basic" {
  role       = aws_iam_role.migrate_goals.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "migrate_goals_dynamo" {
  name = "${var.prefix}-migrate-goals-dynamo"
  role = aws_iam_role.migrate_goals.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = ["dynamodb:Scan", "dynamodb:UpdateItem", "dynamodb:GetItem", "dynamodb:PutItem"]
        Resource = [
          aws_dynamodb_table.goals.arn,
          aws_dynamodb_table.seasons.arn,
          aws_dynamodb_table.teams.arn,
          aws_dynamodb_table.goal_seasons.arn
        ]
      }
    ]
  })
}

resource "aws_lambda_function" "migrate_goals" {
  function_name    = "${var.prefix}-migrate-goals"
  role             = aws_iam_role.migrate_goals.arn
  filename         = data.archive_file.shared_lambda_zip.output_path
  source_code_hash = data.archive_file.shared_lambda_zip.output_base64sha256
  handler          = "bootstrap"
  runtime          = local.lambda_runtime
  timeout          = 300

  environment {
    variables = {
      HANDLER          = "MigrateGoals"
      GOALS_TABLE_NAME = aws_dynamodb_table.goals.name
      SEASONS_TABLE_NAME = aws_dynamodb_table.seasons.name
      TEAMS_TABLE_NAME = aws_dynamodb_table.teams.name
      GOAL_SEASONS_TABLE_NAME = aws_dynamodb_table.goal_seasons.name
    }
  }
}

resource "aws_lambda_invocation" "migrate_goals" {
  function_name = aws_lambda_function.migrate_goals.function_name
  input         = jsonencode({})
  triggers = { version = "1" }
  depends_on    = [aws_lambda_function.migrate_goals]
}

resource "aws_iam_role" "migrate_permissions" {
  name = "${var.prefix}-migrate-permissions"
  assume_role_policy = jsonencode({
    Version = "2012-10-17", Statement = [{Action = "sts:AssumeRole", Effect = "Allow", Principal = { Service = "lambda.amazonaws.com" }}]
  })
}

resource "aws_iam_role_policy_attachment" "migrate_permissions_basic" {
  role       = aws_iam_role.migrate_permissions.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "migrate_permissions_dynamo" {
  name = "${var.prefix}-migrate-permissions-dynamo"
  role = aws_iam_role.migrate_permissions.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = ["dynamodb:Scan", "dynamodb:UpdateItem", "dynamodb:PutItem", "dynamodb:Query", "dynamodb:DeleteItem"]
        Resource = [
          aws_dynamodb_table.role_definitions.arn,
          "${aws_dynamodb_table.role_definitions.arn}/index/*",
          aws_dynamodb_table.ownership_policies.arn,
          "${aws_dynamodb_table.ownership_policies.arn}/index/*"
        ]
      }
    ]
  })
}

resource "aws_lambda_function" "migrate_permissions" {
  function_name    = "${var.prefix}-migrate-permissions"
  role             = aws_iam_role.migrate_permissions.arn
  filename         = data.archive_file.shared_lambda_zip.output_path
  source_code_hash = data.archive_file.shared_lambda_zip.output_base64sha256
  handler          = "bootstrap"
  runtime          = local.lambda_runtime
  timeout          = 300

  environment {
    variables = {
      HANDLER                       = "MigratePermissions"
      ROLE_DEFINITIONS_TABLE_NAME   = aws_dynamodb_table.role_definitions.name
      OWNERSHIP_POLICIES_TABLE_NAME = aws_dynamodb_table.ownership_policies.name
    }
  }
}

resource "aws_lambda_invocation" "migrate_permissions" {
  function_name = aws_lambda_function.migrate_permissions.function_name
  input         = jsonencode({})
  triggers = { version = "1" }
  depends_on    = [aws_lambda_function.migrate_permissions]
}

