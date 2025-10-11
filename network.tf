module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "${var.prefix}-volleygoals"
  cidr = "172.16.0.0/16"
  enable_ipv6 = false

  azs             = ["eu-central-1a", "eu-central-1b", "eu-central-1c"]
  public_subnets  = ["172.16.1.0/24", "172.16.2.0/24", "172.16.3.0/24"]

  enable_nat_gateway = false
  enable_vpn_gateway = false

  tags = merge(
    {
      "Application" = "volleygoals"
    },
    var.tags,
  )
}
