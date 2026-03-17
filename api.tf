resource "aws_api_gateway_rest_api" "api" {
  name        = "${var.prefix}-volleygoals"
  description = "API for VolleyGoals application"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = local.tags
}

resource "aws_api_gateway_resource" "api" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_rest_api.api.root_resource_id
  path_part   = "api"
}

resource "aws_api_gateway_resource" "v1" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.api.id
  path_part   = "v1"
}

resource "aws_api_gateway_authorizer" "this" {
  name          = "${var.prefix}-volleygoals"
  rest_api_id   = aws_api_gateway_rest_api.api.id
  type          = "COGNITO_USER_POOLS"
  provider_arns = [var.cognito_user_pool_arn]
}

resource "aws_acm_certificate" "api" {
  domain_name       = "api.${data.aws_route53_zone.this.name}"
  validation_method = "DNS"
  region            = "eu-central-1" # API Gateway only supports certificates in us-east-1 for edge-optimized and eu-central-1 for regional

  tags = local.tags
}

resource "aws_route53_record" "api_cert_validation" {
  for_each = {
    for dvo in aws_acm_certificate.api.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      record = dvo.resource_record_value
      type   = dvo.resource_record_type
    }
  }

  allow_overwrite = true
  name            = each.value.name
  records         = [each.value.record]
  ttl             = 60
  type            = each.value.type
  zone_id         = data.aws_route53_zone.this.zone_id
}

resource "aws_acm_certificate_validation" "api" {
  certificate_arn         = aws_acm_certificate.api.arn
  validation_record_fqdns = [for record in aws_route53_record.api_cert_validation : record.fqdn]
}

resource "aws_api_gateway_domain_name" "this" {
  regional_certificate_arn = aws_acm_certificate_validation.api.certificate_arn
  domain_name              = "api.${data.aws_route53_zone.this.name}"
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = local.tags
}

resource "aws_route53_record" "api_domain" {
  name    = aws_api_gateway_domain_name.this.domain_name
  type    = "A"
  zone_id = data.aws_route53_zone.this.id

  alias {
    evaluate_target_health = true
    name                   = aws_api_gateway_domain_name.this.regional_domain_name
    zone_id                = aws_api_gateway_domain_name.this.regional_zone_id
  }
}


resource "null_resource" "force_redeploy" {
  triggers = {
    # Redeploy when Lambda source or route configuration changes.
    # Using a stable hash prevents unnecessary replacements (and the cycle they cause)
    # that occurred with the previous timestamp() trigger.
    src_hash = local.go_src_hash
    routes_hash = sha1(join("", [
      for f in sort(fileset(path.module, "routes_*.tf")) :
      filesha1("${path.module}/${f}")
    ]))
  }
}

resource "aws_api_gateway_deployment" "this" {
  rest_api_id = aws_api_gateway_rest_api.api.id

  lifecycle {
    # create_before_destroy removed: it creates a deposed-instance ordering constraint
    # (destroy-old waits for stage update) that cycles back through the lambda modules
    # when those modules are simultaneously transitioning their build resources.
    replace_triggered_by = [
      null_resource.force_redeploy.id
    ]
  }

  depends_on = [
    # Self
    module.get_self_ms,
    module.update_self_ms,
    module.upload_self_picture_ms,
    # Teams
    module.get_team_ms,
    module.get_teams_ms,
    module.create_team_ms,
    module.delete_team_ms,
    module.update_team_ms,
    module.update_team_settings_ms,
    module.get_team_invites_ms,
    module.upload_team_picture_ms,
    module.get_team_activity_ms,
    # Team Members
    module.list_team_members_ms,
    module.add_team_member_ms,
    module.update_team_member_ms,
    module.delete_team_member_ms,
    module.leave_team_ms,
    # Invites
    module.create_invite_ms,
    module.complete_invite_ms,
    module.revoke_invite_ms,
    module.resend_invite_ms,
    module.get_invite_by_token_ms,
    # Users
    module.list_users_ms,
    module.get_user_ms,
    module.delete_user_ms,
    module.update_user_ms,
    # Seasons
    module.create_season_ms,
    module.list_seasons_ms,
    module.get_season_ms,
    module.update_season_ms,
    module.delete_season_ms,
    module.get_season_stats_ms,
    # Goals
    module.create_goal_ms,
    module.list_goals_ms,
    module.get_goal_ms,
    module.update_goal_ms,
    module.delete_goal_ms,
    module.upload_goal_file_ms,
    # Progress Reports
    module.create_progress_report_ms,
    module.list_progress_reports_ms,
    module.get_progress_report_ms,
    module.update_progress_report_ms,
    module.delete_progress_report_ms,
    # Comments
    module.create_comment_ms,
    module.list_comments_ms,
    module.get_comment_ms,
    module.update_comment_ms,
    module.delete_comment_ms,
    module.upload_comment_file_ms,
    # Search & Health
    module.global_search_ms,
    module.health_check_ms,
  ]
}

resource "aws_api_gateway_stage" "this" {
  deployment_id        = aws_api_gateway_deployment.this.id
  rest_api_id          = aws_api_gateway_rest_api.api.id
  stage_name           = "api"
  xray_tracing_enabled = true

  depends_on = [aws_cloudwatch_log_group.api]
}

resource "aws_api_gateway_method_settings" "this" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  stage_name  = aws_api_gateway_stage.this.stage_name
  method_path = "*/*"

  settings {
    metrics_enabled = true
    logging_level   = "INFO"
  }

  depends_on = [
    aws_api_gateway_account.this
  ]
}

resource "aws_api_gateway_base_path_mapping" "this" {
  api_id      = aws_api_gateway_rest_api.api.id
  stage_name  = aws_api_gateway_stage.this.stage_name
  domain_name = aws_api_gateway_domain_name.this.domain_name
}

resource "aws_cloudwatch_log_group" "api" {
  name              = "API-Gateway-Execution-Logs_${aws_api_gateway_rest_api.api.id}/api"
  retention_in_days = 30

  tags = local.tags
}

resource "aws_api_gateway_account" "this" {
  cloudwatch_role_arn = aws_iam_role.cloudwatch.arn
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["apigateway.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "cloudwatch" {
  name               = "api_gateway_cloudwatch_global"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "cloudwatch" {
  statement {
    effect = "Allow"

    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
      "logs:PutLogEvents",
      "logs:GetLogEvents",
      "logs:FilterLogEvents",
    ]

    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "cloudwatch" {
  name   = "default"
  role   = aws_iam_role.cloudwatch.id
  policy = data.aws_iam_policy_document.cloudwatch.json
}
