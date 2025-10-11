data "aws_route53_zone" "this" {
  zone_id = var.dns_zone_id
}
