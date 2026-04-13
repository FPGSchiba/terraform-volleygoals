# DynamoDB Tables
resource "aws_dynamodb_table" "teams" {
  name         = "${var.prefix}-teams"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "tenantId"
    type = "S"
  }

  global_secondary_index {
    name = "tenantIdIndex"
    key_schema {
      attribute_name = "tenantId"
      key_type       = "HASH"
    }
    projection_type = "ALL"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "team_members" {
  name         = "${var.prefix}-team-members"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "teamId"
    type = "S"
  }
  attribute {
    name = "userId"
    type = "S"
  }

  global_secondary_index {
    key_schema {
      attribute_name = "teamId"
      key_type       = "HASH"
    }
    key_schema {
      attribute_name = "userId"
      key_type       = "RANGE"
    }
    name            = "teamUserIdIndex"
    projection_type = "ALL"
  }

  global_secondary_index {
    key_schema {
      attribute_name = "userId"
      key_type       = "HASH"
    }
    name            = "userIdIndex"
    projection_type = "ALL"
  }

  global_secondary_index {
    key_schema {
      attribute_name = "teamId"
      key_type       = "HASH"
    }
    name            = "teamIdIndex"
    projection_type = "ALL"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "invites" {
  name         = "${var.prefix}-invites"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "inviteToken"
    type = "S"
  }

  attribute {
    name = "teamId"
    type = "S"
  }

  global_secondary_index {
    key_schema {
      attribute_name = "inviteToken"
      key_type       = "HASH"
    }
    name            = "tokenIndex"
    projection_type = "ALL"
  }

  global_secondary_index {
    key_schema {
      attribute_name = "teamId"
      key_type       = "HASH"
    }
    name            = "teamIdIndex"
    projection_type = "ALL"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "team_settings" {
  name         = "${var.prefix}-team-settings"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "teamId"
    type = "S"
  }

  global_secondary_index {
    key_schema {
      attribute_name = "teamId"
      key_type       = "HASH"
    }
    name            = "teamIdIndex"
    projection_type = "ALL"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "seasons" {
  name         = "${var.prefix}-seasons"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "goals" {
  name         = "${var.prefix}-goals"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "progress_reports" {
  name         = "${var.prefix}-progress-reports"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "progress" {
  name         = "${var.prefix}-progress"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "comments" {
  name         = "${var.prefix}-comments"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "comment_files" {
  name         = "${var.prefix}-comment-files"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "activities" {
  name         = "${var.prefix}-activities"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "teamId"
    type = "S"
  }

  global_secondary_index {
    name = "teamIdIndex"
    key_schema {
      attribute_name = "teamId"
      key_type       = "HASH"
    }
    projection_type = "ALL"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "goal_seasons" {
  name         = "${var.prefix}-goal-seasons"
  billing_mode = "PAY_PER_REQUEST"
  key_schema {
    attribute_name = "id"
    key_type       = "HASH"
  }

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "goalId"
    type = "S"
  }
  attribute {
    name = "seasonId"
    type = "S"
  }

  global_secondary_index {
    name = "goalIdIndex"
    key_schema {
      attribute_name = "goalId"
      key_type       = "HASH"
    }
    projection_type = "ALL"
  }

  global_secondary_index {
    name = "seasonIdIndex"
    key_schema {
      attribute_name = "seasonId"
      key_type       = "HASH"
    }
    projection_type = "ALL"
  }

  tags = local.tags
}

# Parameter Store
