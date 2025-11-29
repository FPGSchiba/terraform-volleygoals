resource "aws_s3_bucket" "this" {
  bucket_prefix = "${var.prefix}-volleygoals-"
  force_destroy = true

  tags = merge(local.tags, { Name = "${var.prefix}-volleygoals" })
}

resource "aws_s3_bucket_cors_configuration" "this" {
  bucket = aws_s3_bucket.this.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT", "POST"]
    allowed_origins = var.prefix != "dev" ? ["https://${data.aws_route53_zone.this.name}", "https://api.${data.aws_route53_zone.this.name}"] : ["https://${data.aws_route53_zone.this.name}", "https://api.${data.aws_route53_zone.this.name}", "http://localhost:3000"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }

  cors_rule {
    allowed_methods = ["GET"]
    allowed_origins = ["*"]
  }
}

