terraform {
  required_version = ">= 1.10"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5"
    }
    null = {
      source  = "hashicorp/null"
      version = ">= 3"
    }
    http = {
      source  = "hashicorp/http"
      version = "~> 3.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = ">= 2.2"
    }
    uname = {
      source  = "julienlevasseur/uname"
      version = "0.2.3"
    }
    external = {
      source  = "hashicorp/external"
      version = ">= 2.0"
    }
  }
}
