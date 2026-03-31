# Tenants table
resource "aws_dynamodb_table" "tenants" {
  name         = "${var.prefix}-tenants"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  tags = local.tags
}

# Tenant members table
resource "aws_dynamodb_table" "tenant_members" {
  name         = "${var.prefix}-tenant-members"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "tenantId"
    type = "S"
  }
  attribute {
    name = "userId"
    type = "S"
  }

  global_secondary_index {
    name            = "tenantIdIndex"
    hash_key        = "tenantId"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "userIdIndex"
    hash_key        = "userId"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "tenantUserIndex"
    hash_key        = "tenantId"
    range_key       = "userId"
    projection_type = "ALL"
  }

  tags = local.tags
}

# Role definitions table
resource "aws_dynamodb_table" "role_definitions" {
  name         = "${var.prefix}-role-definitions"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "tenantId"
    type = "S"
  }
  attribute {
    name = "name"
    type = "S"
  }

  global_secondary_index {
    name            = "tenantIdIndex"
    hash_key        = "tenantId"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "tenantNameIndex"
    hash_key        = "tenantId"
    range_key       = "name"
    projection_type = "ALL"
  }

  tags = local.tags
}

# Ownership policies table
resource "aws_dynamodb_table" "ownership_policies" {
  name         = "${var.prefix}-ownership-policies"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
  attribute {
    name = "tenantId"
    type = "S"
  }
  attribute {
    name = "resourceType"
    type = "S"
  }

  global_secondary_index {
    name            = "tenantIdIndex"
    hash_key        = "tenantId"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "tenantResourceTypeIndex"
    hash_key        = "tenantId"
    range_key       = "resourceType"
    projection_type = "ALL"
  }

  tags = local.tags
}
