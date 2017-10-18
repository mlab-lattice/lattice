###############################################################################
# Variables

variable "aws_account_id" {}
variable "region" {}

variable "lattice_id" {}
variable "system_id" {}
variable "vpc_id" {}
variable "vpc_cidr_block" {}
variable "build_subnet_ids" {}

variable "build_id" {}
variable "num_instances" {}
variable "instance_type" {}
variable "ami_id" {}
variable "key_name" {}

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
# Build node

module "build_node" {
  source = "/opt/terraform/modules/node/build"

  aws_account_id   = "${var.aws_account_id}"
  lattice_id       = "${var.lattice_id}"
  system_id        = "${var.system_id}"
  build_id         = "${var.build_id}"
  region           = "${var.region}"
  vpc_id           = "${var.vpc_id}"
  vpc_cidr_block   = "${var.vpc_cidr_block}"
  build_subnet_ids = "${var.build_subnet_ids}"
  num_instances    = "${var.num_instances}"
  instance_type    = "${var.instance_type}"
  ami_id           = "${var.ami_id}"
  key_name         = "${var.key_name}"
}
