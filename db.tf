# RDS Database Instance

module "db" {
  source = "github.com/FPGSchiba/terraform-aws-database?ref=v1.1.1"

  name                  = "${var.prefix}-volleygoals"
  vpc_id                = module.vpc.vpc_id
  subnet_ids            = module.vpc.public_subnets
  publicly_accessible   = true
  max_allocated_storage = 50
  security_groups = [
    {
      name        = "${var.prefix}-volleygoals-db"
      description = "Security group for VolleyGoals RDS instance"
      rules = [
        {
          type        = "ingress"
          from_port   = 5432
          to_port     = 5432
          ip_protocol = "tcp"
          ipv4_cidr_blocks = ["172.16.0.0/16", "85.1.208.23/32"] # VPC and home IPv4
          ipv6_cidr_blocks = [] # VPC and home IPv6
        },
        {
          type        = "egress"
          ip_protocol = "-1"
          ipv4_cidr_blocks = ["0.0.0.0/0"]
          ipv6_cidr_blocks = ["::/0"]
        }
      ]
    }
  ]

  tags = merge(
    {
      "Application" = "volleygoals"
    },
    var.tags,
  )
}
