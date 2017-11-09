###############################################################################
# Variables
#

variable "aws_account_id" {}
variable "region" {}

variable "system_id" {}
variable "vpc_id" {}
variable "vpc_cidr_block" {}
variable "build_subnet_ids" {}

variable "build_id" {}
variable "num_instances" {}
variable "instance_type" {}
variable "base_node_ami_id" {}
variable "key_name" {}

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

resource "aws_iam_role" "build_node_role" {
  name               = "${var.system_id}.build-${var.build_id}"
  assume_role_policy = "${module.assume_role_from_ec2_service_policy_doucment.json}"
}

module "assume_role_from_ec2_service_policy_doucment" {
  source = "../../iam/sts/assume-role-from-ec2-service-policy-document"
}

###############################################################################
# Policy

resource "aws_iam_role_policy" "build_node_role_policy" {
  role   = "${aws_iam_role.build_node_role.id}"
  policy = "${data.aws_iam_policy_document.build_node_role_policy_document.json}"
}

data "aws_iam_policy_document" "build_node_role_policy_document" {
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

  # Allow pull from global ecr build repository
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
      "arn:aws:ecr:${var.region}:${var.aws_account_id}:repository/lattice/global/build/*",
    ]
  }

  # Allow push to system ecr repos
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
      "ecr:PutImage",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
    ]

    resources = [
      "arn:aws:ecr:${var.region}:${var.aws_account_id}:repository/lattice/systems/${var.system_id}/*",
    ]
  }
}

###############################################################################
# base node
#

module "base_node" {
  source = "../base"

  system_id = "${var.system_id}"
  name      = "build-${var.build_id}"

  kubelet_labels = "node-role.lattice.mlab.com/build=true"
  kubelet_taints = "node-role.lattice.mlab.com/build=true:NoSchedule"

  region         = "${var.region}"
  vpc_id         = "${var.vpc_id}"
  vpc_cidr_block = "${var.vpc_cidr_block}"
  subnet_ids     = "${var.build_subnet_ids}"
  num_instances  = "${var.num_instances}"
  instance_type  = "${var.instance_type}"
  ami_id         = "${var.base_node_ami_id}"
  key_name       = "${var.key_name}"

  iam_instance_profile_role_name = "${aws_iam_role.build_node_role.name}"
}
