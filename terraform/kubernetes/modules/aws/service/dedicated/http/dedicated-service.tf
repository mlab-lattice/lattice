###############################################################################
# Variables

variable "aws_account_id" {}
variable "region" {}

variable "availability_zones" {
  type = "list"
}

variable "system_id" {}
variable "vpc_id" {}
variable "subnet_ids" {
  type = "list"
}
variable "base_node_ami_id" {}
variable "key_name" {}

variable "service_id" {}
variable "num_instances" {}
variable "instance_type" {}

# The port_numbers variable maps the port that the load balancer should expose to
# the port on the autoscaling group that it should target.

variable "ports" {
  type = "map"
}

###############################################################################
# Provider

provider "aws" {
  region = "${var.region}"
}

###############################################################################
# Service node

module "service_node" {
  source = "../../../node/service"

  aws_account_id = "${var.aws_account_id}"
  region         = "${var.region}"

  system_id  = "${var.system_id}"
  vpc_id     = "${var.vpc_id}"
  subnet_ids = "${var.subnet_ids}"

  service_id       = "${var.service_id}"
  num_instances    = "${var.num_instances}"
  instance_type    = "${var.instance_type}"
  base_node_ami_id = "${var.base_node_ami_id}"
  key_name         = "${var.key_name}"
}

###############################################################################
# ALB

module "load_balancer" {
  source = "../../load-balancer/application"

  region = "${var.region}"

  system_id              = "${var.system_id}"
  vpc_id                 = "${var.vpc_id}"
  autoscaling_group_name = "${module.service_node.autoscaling_group_name}"
  subnet_ids             = "${var.subnet_ids}"

  service_id                = "${var.service_id}"
  service_security_group_id = "${module.service_node.security_group_id}"

  ports = "${var.ports}"
}
