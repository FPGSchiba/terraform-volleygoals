# ─── SNS Topic for Alerts ─────────────────────────────────────────────────────

resource "aws_sns_topic" "alerts" {
  name = "${var.prefix}-volleygoals-alerts"
  tags = local.tags
}

# ─── API Gateway Access Log Group ─────────────────────────────────────────────

resource "aws_cloudwatch_log_group" "api_access_logs" {
  name              = "${var.prefix}-volleygoals-api-access-logs"
  retention_in_days = 90
  tags              = local.tags
}

# ─── CloudWatch Alarms ────────────────────────────────────────────────────────

resource "aws_cloudwatch_metric_alarm" "api_5xx" {
  alarm_name          = "${var.prefix}-api-5xx-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "5XXError"
  namespace           = "AWS/ApiGateway"
  period              = 300
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "API Gateway 5xx errors exceed threshold"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "notBreaching"

  dimensions = {
    ApiName = aws_api_gateway_rest_api.api.name
  }

  tags = local.tags
}

resource "aws_cloudwatch_metric_alarm" "api_4xx" {
  alarm_name          = "${var.prefix}-api-4xx-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "4XXError"
  namespace           = "AWS/ApiGateway"
  period              = 300
  statistic           = "Sum"
  threshold           = 50
  alarm_description   = "API Gateway 4xx errors exceed threshold"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "notBreaching"

  dimensions = {
    ApiName = aws_api_gateway_rest_api.api.name
  }

  tags = local.tags
}

resource "aws_cloudwatch_metric_alarm" "api_latency" {
  alarm_name          = "${var.prefix}-api-p99-latency"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "Latency"
  namespace           = "AWS/ApiGateway"
  period              = 300
  extended_statistic  = "p99"
  threshold           = 5000
  alarm_description   = "API Gateway p99 latency exceeds 5s"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "notBreaching"

  dimensions = {
    ApiName = aws_api_gateway_rest_api.api.name
  }

  tags = local.tags
}

# ─── DynamoDB Throttle Alarms ─────────────────────────────────────────────────

locals {
  dynamodb_tables = {
    teams            = aws_dynamodb_table.teams.name
    team_members     = aws_dynamodb_table.team_members.name
    invites          = aws_dynamodb_table.invites.name
    team_settings    = aws_dynamodb_table.team_settings.name
    seasons          = aws_dynamodb_table.seasons.name
    goals            = aws_dynamodb_table.goals.name
    progress_reports = aws_dynamodb_table.progress_reports.name
    progress         = aws_dynamodb_table.progress.name
    comments         = aws_dynamodb_table.comments.name
    comment_files    = aws_dynamodb_table.comment_files.name
    activities       = aws_dynamodb_table.activities.name
  }

  lambda_function_names = [
    "get-self", "update-self", "upload-self-picture",
    "list-teams", "get-team", "create-team", "update-team", "delete-team",
    "update-team-settings",
    "list-team-members", "add-team-member", "update-team-member", "delete-team-member", "leave-team",
    "upload-team-picture", "get-team-activity", "get-team-invites",
    "create-invite", "complete-invite", "revoke-invite", "resend-invite", "get-invite-by-token",
    "list-users", "get-user", "delete-user", "update-user",
    "create-season", "list-seasons", "get-season", "update-season", "delete-season", "get-season-stats",
    "create-goal", "list-goals", "get-goal", "update-goal", "delete-goal", "upload-goal-file",
    "create-progress-report", "list-progress-reports", "get-progress-report", "update-progress-report", "delete-progress-report",
    "create-comment", "list-comments", "get-comment", "update-comment", "delete-comment", "upload-comment-file",
    "global-search", "health-check",
  ]
}

resource "aws_cloudwatch_metric_alarm" "dynamodb_throttle" {
  for_each = local.dynamodb_tables

  alarm_name          = "${var.prefix}-dynamodb-throttle-${each.key}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ThrottledRequests"
  namespace           = "AWS/DynamoDB"
  period              = 300
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "DynamoDB throttling detected on ${each.value}"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "notBreaching"

  dimensions = {
    TableName = each.value
  }

  tags = local.tags
}

# ─── Log Metric Filters ──────────────────────────────────────────────────────

resource "aws_cloudwatch_log_metric_filter" "lambda_errors" {
  for_each = toset(local.lambda_function_names)

  name           = "${var.prefix}-${each.value}-errors"
  log_group_name = "/aws/lambda/${var.prefix}-${each.value}"
  pattern        = "{ $.level = \"error\" }"

  metric_transformation {
    name          = "LambdaErrorCount"
    namespace     = "${var.prefix}/VolleyGoals"
    value         = "1"
    default_value = "0"
  }
}

resource "aws_cloudwatch_log_metric_filter" "lambda_panics" {
  for_each = toset(local.lambda_function_names)

  name           = "${var.prefix}-${each.value}-panics"
  log_group_name = "/aws/lambda/${var.prefix}-${each.value}"
  pattern        = "{ $.panic = * }"

  metric_transformation {
    name          = "LambdaPanicCount"
    namespace     = "${var.prefix}/VolleyGoals"
    value         = "1"
    default_value = "0"
  }
}

resource "aws_cloudwatch_log_metric_filter" "lambda_access_denied" {
  for_each = toset(local.lambda_function_names)

  name           = "${var.prefix}-${each.value}-access-denied"
  log_group_name = "/aws/lambda/${var.prefix}-${each.value}"
  pattern        = "AccessDenied"

  metric_transformation {
    name          = "LambdaAccessDeniedCount"
    namespace     = "${var.prefix}/VolleyGoals"
    value         = "1"
    default_value = "0"
  }
}

# ─── Alarms on Custom Metrics ─────────────────────────────────────────────────

resource "aws_cloudwatch_metric_alarm" "lambda_error_rate" {
  alarm_name          = "${var.prefix}-lambda-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "LambdaErrorCount"
  namespace           = "${var.prefix}/VolleyGoals"
  period              = 300
  statistic           = "Sum"
  threshold           = 20
  alarm_description   = "Lambda error count exceeds threshold"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "notBreaching"
  tags                = local.tags
}

resource "aws_cloudwatch_metric_alarm" "lambda_panic" {
  alarm_name          = "${var.prefix}-lambda-panics"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "LambdaPanicCount"
  namespace           = "${var.prefix}/VolleyGoals"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "Lambda panic detected"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "notBreaching"
  tags                = local.tags
}

# ─── Unauthorized Access Metric Filter ────────────────────────────────────────

resource "aws_cloudwatch_log_metric_filter" "unauthorized_access" {
  name           = "${var.prefix}-unauthorized-access"
  log_group_name = aws_cloudwatch_log_group.api_access_logs.name
  pattern        = "{ $.status = 401 || $.status = 403 }"

  metric_transformation {
    name          = "UnauthorizedAccessCount"
    namespace     = "${var.prefix}/VolleyGoals"
    value         = "1"
    default_value = "0"
  }
}

resource "aws_cloudwatch_metric_alarm" "unauthorized_access" {
  alarm_name          = "${var.prefix}-unauthorized-access"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "UnauthorizedAccessCount"
  namespace           = "${var.prefix}/VolleyGoals"
  period              = 300
  statistic           = "Sum"
  threshold           = 50
  alarm_description   = "Excessive unauthorized access attempts"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  treat_missing_data  = "notBreaching"
  tags                = local.tags
}

# ─── CloudWatch Dashboard ─────────────────────────────────────────────────────

resource "aws_cloudwatch_dashboard" "main" {
  dashboard_name = "${var.prefix}-volleygoals"

  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          title  = "API Gateway Error Rates"
          region = data.aws_region.current.name
          metrics = [
            ["AWS/ApiGateway", "5XXError", "ApiName", aws_api_gateway_rest_api.api.name, { stat = "Sum", color = "#d62728" }],
            ["AWS/ApiGateway", "4XXError", "ApiName", aws_api_gateway_rest_api.api.name, { stat = "Sum", color = "#ff7f0e" }],
          ]
          period = 300
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          title  = "API Gateway Latency Percentiles"
          region = data.aws_region.current.name
          metrics = [
            ["AWS/ApiGateway", "Latency", "ApiName", aws_api_gateway_rest_api.api.name, { stat = "p50", label = "p50" }],
            ["AWS/ApiGateway", "Latency", "ApiName", aws_api_gateway_rest_api.api.name, { stat = "p90", label = "p90" }],
            ["AWS/ApiGateway", "Latency", "ApiName", aws_api_gateway_rest_api.api.name, { stat = "p99", label = "p99", color = "#d62728" }],
          ]
          period = 300
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 12
        height = 6
        properties = {
          title  = "Lambda Errors (aggregated)"
          region = data.aws_region.current.name
          metrics = [
            [{ expression = "SEARCH('{AWS/Lambda,FunctionName} MetricName=\"Errors\" FunctionName=\"${var.prefix}-\"', 'Sum', 300)", id = "errors", label = "Lambda Errors" }],
          ]
          period = 300
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 6
        width  = 12
        height = 6
        properties = {
          title  = "DynamoDB Throttling"
          region = data.aws_region.current.name
          metrics = [
            for key, name in local.dynamodb_tables :
            ["AWS/DynamoDB", "ThrottledRequests", "TableName", name, { stat = "Sum", label = key }]
          ]
          period = 300
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 12
        width  = 12
        height = 6
        properties = {
          title  = "Custom: Lambda Errors & Panics"
          region = data.aws_region.current.name
          metrics = [
            ["${var.prefix}/VolleyGoals", "LambdaErrorCount", { stat = "Sum", color = "#d62728" }],
            ["${var.prefix}/VolleyGoals", "LambdaPanicCount", { stat = "Sum", color = "#9467bd" }],
            ["${var.prefix}/VolleyGoals", "LambdaAccessDeniedCount", { stat = "Sum", color = "#ff7f0e" }],
          ]
          period = 300
          view   = "timeSeries"
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 12
        width  = 12
        height = 6
        properties = {
          title  = "Unauthorized Access Attempts"
          region = data.aws_region.current.name
          metrics = [
            ["${var.prefix}/VolleyGoals", "UnauthorizedAccessCount", { stat = "Sum", color = "#d62728" }],
          ]
          period = 300
          view   = "timeSeries"
        }
      },
    ]
  })
}
