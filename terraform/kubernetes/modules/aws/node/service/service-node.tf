###############################################################################
# Variables
#

variable "aws_account_id" {}
variable "region" {}

variable "system_id" {}
variable "vpc_id" {}
variable "subnet_ids" {}
variable "base_node_ami_id" {}
variable "key_name" {}

variable "service_id" {}
variable "num_instances" {}
variable "instance_type" {}

###############################################################################
# Output
#

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

resource "aws_iam_role" "service_node_role" {
  name = "${var.system_id}.service-${var.service_id}"

  //  name               = "${var.lattice_id}.${var.system_id}.service-${var.service_id}"
  assume_role_policy = "${module.assume_role_from_ec2_service_policy_doucment.json}"
}

module "assume_role_from_ec2_service_policy_doucment" {
  source = "../../iam/sts/assume-role-from-ec2-service-policy-document"
}

###############################################################################
# Policy

resource "aws_iam_role_policy" "service_node_role_policy" {
  role   = "${aws_iam_role.service_node_role.id}"
  policy = "${data.aws_iam_policy_document.service_node_role_policy_document.json}"
}

data "aws_iam_policy_document" "service_node_role_policy_document" {
  # Allow ec2 read-only
  statement {
    effect = "Allow"

    actions = [
      "ec2:Describe*",
    ]

    resources = [
      "*",
    ]
  }

  # Allow ecr get-authorization-token
  statement {
    effect = "Allow"

    actions = [
      "ecr:GetAuthorizationToken",
    ]

    resources = [
      "*",
    ]
  }

  # Allow pull from system repos
  statement {
    effect = "Allow"

    actions = [
      "ecr:GetAuthorizationToken",
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetDownloadUrlForLayer",
      "ecr:GetRepositoryPolicy",
      "ecr:DescribeRepositories",
      "ecr:ListImages",
      "ecr:BatchGetImage",
    ]

    resources = [
      "arn:aws:ecr:${var.region}:${var.aws_account_id}:repository/component-builds",
    ]
  }
}

###############################################################################
# base node
#

module "base_node" {
  source = "../base"

  system_id = "${var.system_id}"
  name      = "service-${var.service_id}"

  kubelet_labels = "lattice/service=${var.service_id}"
  kubelet_taints = "lattice/service=${var.service_id}:NoSchedule"

  region        = "${var.region}"
  vpc_id        = "${var.vpc_id}"
  subnet_ids    = "${var.subnet_ids}"
  num_instances = "${var.num_instances}"
  instance_type = "${var.instance_type}"
  ami_id        = "${var.base_node_ami_id}"
  key_name      = "${var.key_name}"

  iam_instance_profile_role_name = "${aws_iam_role.service_node_role.name}"
}
