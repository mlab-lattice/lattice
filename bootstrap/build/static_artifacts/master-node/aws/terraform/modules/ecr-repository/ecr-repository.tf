###############################################################################
# Variables

variable "region" {}
variable "repository" {}

###############################################################################
# Backend

terraform {
  backend "s3" {
    encrypt = true
  }
}

###############################################################################
# Provider

provider "aws" {
  region = "${var.region}"
}

###############################################################################
# ecr

resource "aws_ecr_repository" "repository" {
  name = "${var.repository}"
}
