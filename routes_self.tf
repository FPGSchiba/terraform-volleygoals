# Self picture presign

resource "aws_api_gateway_resource" "self_picture" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.self.id
  path_part   = "picture"
}

resource "aws_api_gateway_resource" "self_picture_presign" {
  rest_api_id = aws_api_gateway_rest_api.api.id
  parent_id   = aws_api_gateway_resource.self_picture.id
  path_part   = "presign"
}

module "upload_self_picture_ms" {
  source = "github.com/FPGSchiba/terraform-aws-microservice?ref=v2.4.0"

  api_id                = aws_api_gateway_rest_api.api.id
  code_dir              = "${path.module}/files/src"
  cors_enabled          = true
  control_allow_origin  = local.cors_allowed_origin
  http_methods          = ["GET"]
  name_overwrite        = "upload-self-picture"
  path_name             = "presign"
  create_resource       = false
  existing_resource_id  = aws_api_gateway_resource.self_picture_presign.id
  prefix                = var.prefix
  authorizer_id         = aws_api_gateway_authorizer.this.id
  authorization_type    = "COGNITO_USER_POOLS"
  enable_tracing        = true
  timeout               = 29
  vpc_networked         = false
  environment_variables = local.lambda_environment_variables
  tags                  = local.tags
  layer_arns            = local.lambda_layer_arns
  json_logging          = true
  handler_name          = "UploadSelfPicture"
  pre_built_zip         = data.archive_file.shared_lambda_zip.output_path

  additional_iam_statements = [
    {
      actions   = ["s3:PutObject"]
      resources = ["${aws_s3_bucket.this.arn}/users/*"]
    },
    {
      actions   = ["cognito-idp:AdminUpdateUserAttributes", "cognito-idp:AdminGetUser", "cognito-idp:AdminListGroupsForUser"]
      resources = [var.cognito_user_pool_arn]
    },
  ]

  depends_on = [
    aws_api_gateway_rest_api.api,
    aws_api_gateway_resource.self_picture_presign,
    data.archive_file.shared_lambda_zip,
  ]
}
