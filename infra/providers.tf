terraform {
  required_version = "~> 1.4"

  required_providers {
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }

    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.66"
    }
  }

  backend "s3" {
    region         = "eu-north-1"
    bucket         = "abborre-terraform-state"
    key            = "terraform.tfstate"
    dynamodb_table = "abborre-terraform-state-lock"
  }
}

provider "aws" {
  region = "eu-north-1"

  default_tags {
    tags = {
      Workspace = terraform.workspace
      App       = "abborre"
    }
  }
}
