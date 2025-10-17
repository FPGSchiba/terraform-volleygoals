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

# Parameter Store
