# DynamoDB Tables
resource "aws_dynamodb_table" "teams" {
  name         = "${var.prefix}-teams"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "team_members" { // Maybe not needed
  name         = "${var.prefix}-team-members"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "invites" {
  name         = "${var.prefix}-invites"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "team_settings" {
  name         = "${var.prefix}-team-settings"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "seasons" {
  name         = "${var.prefix}-seasons"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "goals" {
  name         = "${var.prefix}-goals"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "progress_reports" {
  name         = "${var.prefix}-progress-reports"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "progress" {
  name         = "${var.prefix}-progress"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "comments" {
  name         = "${var.prefix}-comments"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

resource "aws_dynamodb_table" "comment_files" {
  name         = "${var.prefix}-comment-files"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

# Parameter Store
