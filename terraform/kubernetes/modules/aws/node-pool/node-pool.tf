###############################################################################
# Variables
#

variable "aws_account_id" {}
variable "region" {}

variable "lattice_id" {}
variable "vpc_id" {}
variable "subnet_ids" {}
variable "master_node_security_group_id" {}
variable "worker_node_ami_id" {}
variable "key_name" {}

variable "name" {}
variable "num_instances" {}
variable "instance_type" {}

variable "kubelet_port" {
  default = 10250
}

variable "kube_bootstrap_token" {}
variable "kube_apiserver_private_ip" {}
variable "kube_apiserver_port" {}

###############################################################################
# Output
#

output "autoscaling_group_id" {
  value = "${module.node.autoscaling_group_id}"
}

output "autoscaling_group_name" {
  value = "${module.node.autoscaling_group_name}"
}

output "autoscaling_group_desired_capacity" {
  value = "${module.node.autoscaling_group_desired_capacity}"
}

output "security_group_id" {
  value = "${module.node.security_group_id}"
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

resource "aws_iam_role" "node_pool_role" {
  assume_role_policy = "${module.assume_role_from_ec2_service_policy_doucment.json}"
}

module "assume_role_from_ec2_service_policy_doucment" {
  source = "../iam/sts/assume-role-from-ec2-service-policy-document"
}

###############################################################################
# Policy

resource "aws_iam_role_policy" "service_node_role_policy" {
  role   = "${aws_iam_role.node_pool_role.id}"
  policy = "${data.aws_iam_policy_document.node_pool_role_policy_document.json}"
}

data "aws_iam_policy_document" "node_pool_role_policy_document" {
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

  # Allow pull from component builds
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
      "arn:aws:ecr:${var.region}:${var.aws_account_id}:repository/${var.lattice_id}.component-builds",
    ]
  }
}

###############################################################################
# worker node
#

module "node" {
  source = "../node/base"

  lattice_id = "${var.lattice_id}"
  name       = "node-pool-${var.name}"

  kubelet_labels = "node-pool.lattice.mlab.com/id=${var.name}"
  kubelet_taints = "node-pool.lattice.mlab.com/id=${var.name}:NoSchedule"

  region        = "${var.region}"
  vpc_id        = "${var.vpc_id}"
  subnet_ids    = "${var.subnet_ids}"
  num_instances = "${var.num_instances}"
  instance_type = "${var.instance_type}"
  ami_id        = "${var.worker_node_ami_id}"
  key_name      = "${var.key_name}"

  //  kube_bootstrap_token      = "${var.kube_bootstrap_token}"
  //  kube_apiserver_private_ip = "${var.kube_apiserver_private_ip}"
  //  kube_apiserver_port       = "${var.kube_apiserver_port}"

  iam_instance_profile_role_name = "${aws_iam_role.node_pool_role.name}"
}

###############################################################################
# Security Group

resource "aws_security_group_rule" "allow_kubelet_from_master" {
  security_group_id = "${module.node.security_group_id}"

  protocol                 = "tcp"
  from_port                = "${var.kubelet_port}"
  to_port                  = "${var.kubelet_port}"
  type                     = "ingress"
  source_security_group_id = "${var.master_node_security_group_id}"
}
