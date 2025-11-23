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

variable "cognito_user_pool_arn" {
  description = "List of Cognito User Pool ARNs for API Gateway authorizer"
  type        = string
}

variable "dns_zone_id" {
  description = "Route53 Hosted Zone ID for DNS records"
  type        = string
}

variable "ses_tenant_name" {
  description = "Tenant name for multi-tenant setup"
  type        = string
  default     = "default_tenant"
}

variable "ses_configuration_set_name" {
  description = "SES Configuration Set name for email sending"
  type        = string
  default     = "default_configuration_set"

}
