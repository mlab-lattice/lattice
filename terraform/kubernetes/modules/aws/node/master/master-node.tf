###############################################################################
# Variables
#

variable "aws_account_id" {}
variable "region" {}

variable "system_id" {}
variable "system_definition_url" {}
variable "system_s3_bucket" {}
variable "vpc_id" {}
variable "subnet_id" {}
variable "subnet_ids" {}
variable "base_node_ami_id" {}
variable "route53_private_zone_id" {}

variable "name" {}
variable "instance_type" {}
variable "ami_id" {}
variable "key_name" {}

###############################################################################
# Data
#

data "aws_vpc" "vpc" {
  id = "${var.vpc_id}"
}

data "aws_subnet" "master_subnet" {
  id = "${var.subnet_id}"
}

###############################################################################
# Output
#

output "autoscaling_group_id" {
  value = "${module.base_node.autoscaling_group_id}"
}

output "autoscaling_group_name" {
  value = "${module.base_node.autoscaling_group_name}"
}

output "security_group_id" {
  value = "${module.base_node.security_group_id}"
}

###############################################################################
# Provider
#

provider "aws" {
  region = "${var.region}"
}

###############################################################################
# IAM
#

###############################################################################
# Role

resource "aws_iam_role" "master_node_role" {
  name               = "${var.system_id}.master-${var.name}"
  assume_role_policy = "${module.assume_role_from_ec2_service_policy_doucment.json}"
}

module "assume_role_from_ec2_service_policy_doucment" {
  source = "../../iam/sts/assume-role-from-ec2-service-policy-document"
}

###############################################################################
# Policy

resource "aws_iam_role_policy" "master_node_role_policy" {
  role   = "${aws_iam_role.master_node_role.id}"
  policy = "${data.aws_iam_policy_document.master_node_role_policy_document.json}"
}

data "aws_iam_policy_document" "master_node_role_policy_document" {
  # Allow all autoscaling
  statement {
    effect = "Allow"

    actions = [
      "autoscaling:*",
    ]

    resources = [
      "*",
    ]
  }

  # Allow all ec2
  statement {
    effect = "Allow"

    actions = [
      "ec2:*",
    ]

    resources = [
      "*",
    ]
  }

  # Allow all elb
  statement {
    effect = "Allow"

    actions = [
      "elasticloadbalancing:*",
    ]

    resources = [
      "*",
    ]
  }

  # Allow all ecr
  statement {
    effect = "Allow"

    actions = [
      "ecr:*",
    ]

    resources = [
      "*",
    ]
  }

  # Allow all iam
  statement {
    effect = "Allow"

    actions = [
      "iam:*",
    ]

    resources = [
      "*",
    ]
  }

  # Allow all route53
  statement {
    effect = "Allow"

    actions = [
      "route53:*",
    ]

    resources = [
      "*",
    ]
  }

  # Allow all s3
  statement {
    effect = "Allow"

    actions = [
      "s3:*",
    ]

    resources = [
      "*",
    ]
  }
}

###############################################################################
# base node
#

module "base_node" {
  source = "../base"

  system_id = "${var.system_id}"
  name      = "master-${var.name}"

  kubelet_labels = "node-role.kubernetes.io/master=true,node-role.lattice.mlab.com/master=true"
  kubelet_taints = "node-role.lattice.mlab.com/master=true:NoSchedule"

  additional_user_data = <<USER_DATA
{
  "aws_account_id": "${var.aws_account_id}",
  "system_id": "${var.system_id}",
  "system_definition_url": "${var.system_definition_url}",
  "name": "${var.name}",
  "base_node_ami_id": "${var.base_node_ami_id}",
  "subnet_ids": "${var.subnet_ids}",
  "key_name": "${var.key_name}",
  "route53_private_zone_id": "${var.route53_private_zone_id}",
  "state_s3_bucket": "${var.system_s3_bucket}",
  "state_s3_key_prefix": "masters/nodes/${var.name}/state",
  "vpc_id": "${var.vpc_id}",
  "vpc_cidr_block": "${data.aws_vpc.vpc.cidr_block}",
  "master_node_security_group_id": "${module.base_node.security_group_id}"
}
USER_DATA

  region        = "${var.region}"
  vpc_id        = "${var.vpc_id}"
  subnet_ids    = "${var.subnet_id}"
  num_instances = 1
  instance_type = "${var.instance_type}"
  ami_id        = "${var.ami_id}"
  key_name      = "${var.key_name}"

  iam_instance_profile_role_name = "${aws_iam_role.master_node_role.name}"
}

###############################################################################
# EBS
#

resource "aws_ebs_volume" "master_node_etcd_volume" {
  availability_zone = "${data.aws_subnet.master_subnet.availability_zone}"
  size              = 10
  encrypted         = true

  tags {
    KubernetesCluster = "lattice.system.${var.system_id}"
    Name              = "lattice.system.${var.system_id}.master-${var.name}-etcd"
  }
}
