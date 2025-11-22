locals {
  tags = merge(
    {
      "Application" = "volleygoals"
    },
    var.tags,
  )
  lambda_environment_variables = {
    "TEAMS_TABLE_NAME"            = aws_dynamodb_table.teams.name
    "TEAM_MEMBERS_TABLE_NAME"     = aws_dynamodb_table.team_members.name
    "INVITE_TABLE_NAME"           = aws_dynamodb_table.invites.name
    "TEAM_SETTINGS_TABLE_NAME"    = aws_dynamodb_table.team_settings.name
    "SEASONS_TABLE_NAME"          = aws_dynamodb_table.seasons.name
    "GOALS_TABLE_NAME"            = aws_dynamodb_table.goals.name
    "PROGRESS_REPORTS_TABLE_NAME" = aws_dynamodb_table.progress_reports.name
    "PROGRESS_TABLE_NAME"         = aws_dynamodb_table.progress.name
    "COMMENTS_TABLE_NAME"         = aws_dynamodb_table.comments.name
    "COMMENT_FILES_TABLE_NAME"    = aws_dynamodb_table.comment_files.name
    "OTEL_PROPAGATORS"            = "xray"
    "OTEL_SERVICE_NAME"           = "volleygoals"
    "OTEL_TRACES_SAMPLER"         = "always_on"
    "OTEL_RESOURCE_ATTRIBUTES"    = "service.name=volleygoals"
    "EMAIL_SENDER"                = "no-reply@${data.aws_route53_zone.this.name}"
    "TENANT_NAME"                 = var.ses_tenant_name
    "CONFIGURATION_SET_NAME"      = var.ses_configuration_set_name
    "FRONTEND_BASE_URL"           = "https://${data.aws_route53_zone.this.name}"
  }
  lambda_layer_arns = [
    "arn:aws:lambda:${data.aws_region.current.region}:901920570463:layer:aws-otel-collector-amd64-ver-0-117-0:1" # Me hates it, as it is hardcoded
  ]
}
