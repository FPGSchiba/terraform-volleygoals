##########################
# S3 bucket + CloudFront
# - S3 bucket for user documents/profile pictures
# - BucketOwnerEnforced to make bucket owner authoritative
# - Public access blocked
# - CloudFront Origin Access Control (OAC) with SigV4 signing
# - CloudFront distribution with origin_access_control_id
# - Bucket policy: allow CloudFront (aws:SourceArn) and allow principals from this AWS account
##########################

data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "this" {
  bucket_prefix = "${var.prefix}-volleygoals-"
  force_destroy = true

  tags = merge(local.tags, { Name = "${var.prefix}-volleygoals" })
}

resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.this.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_ownership_controls" "this" {
  bucket = aws_s3_bucket.this.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
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
    allowed_methods = ["GET", "HEAD"]
    allowed_headers = ["*"]
    allowed_origins = ["*"]
  }
}

locals {
  # Policy grants:
  # 1) CloudFront OAC requests (signed via SigV4) from this distribution
  # 2) Any principal whose account equals our account (matches presigned URLs and temporary credentials)
  origin_bucket_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowCloudFrontOACReadOnly"
        Effect    = "Allow"
        Principal = { Service = "cloudfront.amazonaws.com" }
        Action    = ["s3:GetObject"]
        Resource  = ["${aws_s3_bucket.this.arn}/*"]
        Condition = {
          StringEquals = {
            "aws:SourceArn" = aws_cloudfront_distribution.cdn.arn
          }
        }
      },
      {
        Sid       = "AllowAccountPrincipalsGetAndPutObject"
        Effect    = "Allow"
        Principal = { AWS = "*" }
        Action    = ["s3:GetObject", "s3:PutObject"]
        Resource  = ["${aws_s3_bucket.this.arn}/*"]
        Condition = {
          StringEquals = {
            "aws:PrincipalAccount" = data.aws_caller_identity.current.account_id
          }
        }
      }
    ]
  })
}

resource "aws_s3_bucket_policy" "this" {
  bucket = aws_s3_bucket.this.id
  policy = local.origin_bucket_policy
}

# CloudFront Origin Access Control (OAC) for S3 origin using SigV4
resource "aws_cloudfront_origin_access_control" "oac" {
  name                              = "oac-for-${var.prefix}-s3-origin"
  description                       = "OAC for ${var.prefix} S3 origin"
  signing_protocol                  = "sigv4"
  signing_behavior                  = "always"
  origin_access_control_origin_type = "s3"
}

# CloudFront distribution
resource "aws_cloudfront_distribution" "cdn" {
  enabled = true

  aliases = ["cdn.${data.aws_route53_zone.this.name}"]

  origin {
    domain_name              = aws_s3_bucket.this.bucket_regional_domain_name
    origin_id                = "${var.prefix}-s3-origin"
    origin_access_control_id = aws_cloudfront_origin_access_control.oac.id
  }

  default_cache_behavior {
    target_origin_id = "${var.prefix}-s3-origin"

    allowed_methods = ["GET", "HEAD", "OPTIONS"]
    cached_methods  = ["GET", "HEAD"]

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
      headers = ["Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"]
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  # path-specific cache for avatars
  ordered_cache_behavior {
    path_pattern     = "users/*"
    target_origin_id = "${var.prefix}-s3-origin"

    allowed_methods = ["GET", "HEAD", "OPTIONS"]
    cached_methods  = ["GET", "HEAD"]

    forwarded_values {
      query_string = false
      cookies { forward = "none" }
      headers = ["Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"]
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate.cdn_cert.arn
    ssl_support_method  = "sni-only"
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  tags = local.tags
}

resource "aws_route53_record" "cdn_names" {
  for_each = aws_cloudfront_distribution.cdn.aliases
  zone_id  = data.aws_route53_zone.this.zone_id
  name     = each.value
  type     = "A"

  alias {
    name                   = aws_cloudfront_distribution.cdn.domain_name
    zone_id                = aws_cloudfront_distribution.cdn.hosted_zone_id
    evaluate_target_health = false
  }
}

# ACM cert + validation
resource "aws_acm_certificate" "cdn_cert" {
  domain_name       = "cdn.${data.aws_route53_zone.this.name}"
  validation_method = "DNS"
  region            = "us-east-1"

  lifecycle {
    create_before_destroy = true
  }

  tags = local.tags
}

resource "aws_route53_record" "cdn_validation" {
  for_each = {
    for dvo in aws_acm_certificate.cdn_cert.domain_validation_options : dvo.domain_name => {
      name   = dvo.resource_record_name
      type   = dvo.resource_record_type
      record = dvo.resource_record_value
    }
  }

  zone_id = data.aws_route53_zone.this.zone_id
  name    = each.value.name
  type    = each.value.type
  ttl     = 600
  records = [each.value.record]
}
