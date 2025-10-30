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
  }
}
