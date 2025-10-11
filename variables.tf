variable "tags" {
  description = "A map of tags to add to all resources"
  type        = map(string)
  default     = {}
}

variable "prefix" {
  description = "Prefix to use for resource names"
  type        = string
  default     = "dev"
}

variable "cognito_user_pool_arns" {
  description = "List of Cognito User Pool ARNs for API Gateway authorizer"
  type        = list(string)
}

variable "dns_zone_id" {
  description = "Route53 Hosted Zone ID for DNS records"
  type        = string
}
