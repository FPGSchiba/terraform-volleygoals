# Self
resource "aws_api_gateway_resource" "self" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "self"
}

module "get_self_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "get-self"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.self.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "GetSelf"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Query",
      ]
      resources = [
        aws_dynamodb_table.team_settings.arn,
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.self,
  ]
}

module "update_self_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["PATCH"]
  name_overwrite        = "update-self"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.self.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "UpdateSelf"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Query",
        "dynamodb:UpdateItem",
      ]
      resources = [
        aws_dynamodb_table.team_settings.arn,
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.self,
  ]
}

# Teams
resource "aws_api_gateway_resource" "teams" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.v1.id
  path_part   = "teams"
}

resource "aws_api_gateway_resource" "teams_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams.id
  path_part   = "{teamId}"
}

module "get_teams_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "list-teams"
  path_name             = "teams"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "ListTeams"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Scan",
      ]
      resources = [
        aws_dynamodb_table.teams.arn
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams
  ]
}

module "get_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "get-team"
  path_name             = "{teamId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_id.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "GetTeam"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:GetItem",
      ]
      resources = [
        aws_dynamodb_table.teams.arn
      ]
    },
    {
      actions = [
        "dynamodb:Query",
      ]
      resources = [
        "${aws_dynamodb_table.team_settings.arn}/index/teamIdIndex",
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams_id
  ]
}

module "create_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  go_build_tags         = ["createTeam"]
  cors_enabled          = false
  http_methods          = ["POST"]
  name_overwrite        = "create-team"
  path_name             = "teams"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "CreateTeam"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:PutItem",
        "dynamodb:Scan",
      ]
      resources = [
        aws_dynamodb_table.teams.arn,
      ]
    },
    {
      actions = [
        "dynamodb:PutItem",
      ]
      resources = [
        aws_dynamodb_table.team_settings.arn,
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams,
    module.get_teams_ms
  ]
}

module "delete_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-team"
  path_name             = "{teamId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_id.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "DeleteTeam"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:PutItem",
      ]
      resources = [
        aws_dynamodb_table.teams.arn,
        "${aws_dynamodb_table.team_members.arn}/*",
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams_id,
    module.get_team_ms
  ]
}

module "update_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["PATCH"]
  name_overwrite        = "update-team"
  path_name             = "{teamId}"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.teams_id.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "UpdateTeam"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:PutItem",
        "dynamodb:GetItem",
      ]
      resources = [
        aws_dynamodb_table.teams.arn,
        "${aws_dynamodb_table.teams.arn}/*",
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.teams_id,
    module.get_team_ms
  ]
}

# Team Settings

resource "aws_api_gateway_resource" "team_settings" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_id.id
  path_part   = "settings"
}

module "update_team_settings_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["PATCH"]
  name_overwrite        = "update-team-settings"
  path_name             = "settings"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_settings.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "UpdateTeamSettings"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Query",
        "dynamodb:UpdateItem",
      ]
      resources = [
        aws_dynamodb_table.team_settings.arn,
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_settings,
  ]
}

# Team Members

resource "aws_api_gateway_resource" "team_members" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.teams_id.id
  path_part   = "members"
}

resource "aws_api_gateway_resource" "team_member_id" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.team_members.id
  path_part   = "{memberId}"
}

module "list_team_members_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["GET"]
  name_overwrite        = "list-team-members"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_members.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "ListTeamMembers"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Query",
        "dynamodb:UpdateItem",
      ]
      resources = [
        aws_dynamodb_table.team_settings.arn,
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_members,
  ]
}

module "add_team_member_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["POST"]
  name_overwrite        = "add-team-member"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_members.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "AddTeamMember"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Query",
        "dynamodb:UpdateItem",
      ]
      resources = [
        aws_dynamodb_table.team_settings.arn,
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_members,
  ]
}

module "update_team_member_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  http_methods          = ["PATCH"]
  name_overwrite        = "update-team-member"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_member_id.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "UpdateTeamMember"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Query",
        "dynamodb:UpdateItem",
      ]
      resources = [
        aws_dynamodb_table.team_settings.arn,
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_settings,
  ]
}

module "delete_team_member_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["DELETE"]
  name_overwrite        = "delete-team-member"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_member_id.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "RemoveTeamMember"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Query",
        "dynamodb:UpdateItem",
      ]
      resources = [
        aws_dynamodb_table.team_settings.arn,
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_settings,
  ]
}

module "leave_team_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.3.9"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = false
  http_methods          = ["DELETE"]
  name_overwrite        = "leave-team"
  path_name             = "members"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.team_members.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns

  go_additional_ldflags = {
    "github.com/fpgschiba/volleygoals/router.SelectedHandler" = "LeaveTeam"
  }

  additional_iam_statements = [
    {
      actions = [
        "dynamodb:Query",
        "dynamodb:UpdateItem",
      ]
      resources = [
        aws_dynamodb_table.team_settings.arn,
      ]
    }
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.team_settings,
  ]
}
