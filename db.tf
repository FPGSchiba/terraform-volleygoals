# DynamoDB Tables

resource "aws_dynamodb_table" "test" {
  name         = "${var.prefix}-test"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }

  tags = merge(
    {
      "Application" = "volleygoals"
    },
    var.tags,
  )
}

resource "aws_dynamodb_table" "teams" {
  name         = "${var.prefix}-teams"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "Id"
  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "name"
    type = "S"
  }
  attribute {
    name = "status" // active, inactive
    type = "S"
  }
  attribute {
    name = "createdAt"
    type = "S" // ISO 8601
  }
  attribute {
    name = "updatedAt"
    type = "S" // ISO 8601
  }
  attribute {
    name = "deletedAt"
    type = "S" // ISO 8601
  }

  tags = merge(
    {
      "Application" = "volleygoals"
    },
    var.tags,
  )
}

# Parameter Store
