###############################################################################
# Variables

variable "region" {}

variable "system_id" {}
variable "vpc_id" {}
variable "subnet_ids" {}

variable "name" {}
variable "num_instances" {}
variable "instance_type" {}
variable "ami_id" {}
variable "key_name" {}
variable "iam_instance_profile_role_name" {}

variable "additional_user_data" {
  type    = "string"
  default = "{}"
}

variable "kubelet_labels" {}
variable "kubelet_taints" {}

###############################################################################
# Output

output "autoscaling_group_id" {
  value = "${aws_autoscaling_group.node_autoscaling_group.id}"
}

output "autoscaling_group_name" {
  value = "${aws_autoscaling_group.node_autoscaling_group.name}"
}

output "security_group_id" {
  value = "${aws_security_group.node_auto_scaling_group.id}"
}

###############################################################################
# Data

data "aws_vpc" "vpc" {
  id = "${var.vpc_id}"
}

###############################################################################
# Provider

provider "aws" {
  region = "${var.region}"
}

###############################################################################
# IAM

# instance profile
resource "aws_iam_instance_profile" "iam_instance_profile" {
  name = "lattice.system.${var.system_id}.${var.name}"
  role = "${var.iam_instance_profile_role_name}"
}

###############################################################################
# Security Groups

# security group
resource "aws_security_group" "node_auto_scaling_group" {
  name = "lattice.system.${var.system_id}.${var.name}"

  vpc_id = "${var.vpc_id}"

  lifecycle {
    create_before_destroy = true
  }

  tags {
    KubernetesCluster = "lattice.system.${var.system_id}"
    Name              = "lattice.system.${var.system_id}.${var.name}"
  }
}

# FIXME: probably eventually don't want this by default
# Allow all egress traffic
resource "aws_security_group_rule" "auto_scalling_group_allow_egress" {
  security_group_id = "${aws_security_group.node_auto_scaling_group.id}"

  type        = "egress"
  from_port   = 0
  to_port     = 0
  protocol    = -1
  cidr_blocks = ["0.0.0.0/0"]

  lifecycle {
    create_before_destroy = true
  }
}

# Allow flannel vxlan udp traffic
resource "aws_security_group_rule" "auto_scalling_group_allow_ingress_flannel_vxlan" {
  security_group_id = "${aws_security_group.node_auto_scaling_group.id}"

  type        = "ingress"
  from_port   = 8472
  to_port     = 8472
  protocol    = "udp"
  cidr_blocks = ["${data.aws_vpc.vpc.cidr_block}"]

  lifecycle {
    create_before_destroy = true
  }
}

# FIXME: add rule allowing traffic from master node to kubelet port (10250) for logs
# TODO: find out why it's 10250

# FIXME: TEMPORARY FOR TESTING
resource "aws_security_group" "temporary_ssh_group" {
  name = "lattice-system-${var.system_id}-${var.name} TEMPORARY SSH RULE"

  vpc_id = "${var.vpc_id}"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  lifecycle {
    create_before_destroy = true
  }

  tags {
    KubernetesCluster = "lattice.system.${var.system_id}"
    Name              = "lattice.system.${var.system_id}.${var.name}-TEMP-SSH"
  }
}

###############################################################################
# Autoscaling Groups

# launch configuration
resource "aws_launch_configuration" "aws_launch_configuration" {
  image_id      = "${var.ami_id}"
  instance_type = "${var.instance_type}"
  key_name      = "${var.key_name}"

  iam_instance_profile = "${aws_iam_instance_profile.iam_instance_profile.name}"

  user_data = <<EOF
{
  "ignition": {
    "version": "2.0.0",
    "config": {}
  },
  "storage": {},
  "systemd": {
    "units": [
      {
        "name": "kubelet.service",
        "dropins": [
          {
            "name": "10-override.conf",
            "contents": "[Service]\nEnvironment=\"KUBELET_LABELS=${var.kubelet_labels}\"\nEnvironment=\"KUBELET_TAINTS=${var.kubelet_taints}\""
          }
        ]
      }
    ]
  },
  "networkd": {},
  "passwd": {},
  "lattice": ${var.additional_user_data}
}
EOF

  # TODO: remove temporary_ssh_group when done testing
  security_groups = [
    "${aws_security_group.node_auto_scaling_group.id}",
    "${aws_security_group.temporary_ssh_group.id}",
  ]

  # Needed to be able to talk to public internet for build deps
  associate_public_ip_address = true

  lifecycle {
    create_before_destroy = true
  }
}

# autoscaling group
resource "aws_autoscaling_group" "node_autoscaling_group" {
  name                 = "lattice.system.${var.system_id}.${var.name}-${aws_launch_configuration.aws_launch_configuration.name}"
  launch_configuration = "${aws_launch_configuration.aws_launch_configuration.name}"

  desired_capacity = "${var.num_instances}"
  min_size         = "${var.num_instances}"
  max_size         = "${var.num_instances}"

  vpc_zone_identifier = ["${split(",", var.subnet_ids)}"]

  lifecycle {
    create_before_destroy = true
  }

  tag {
    key   = "KubernetesCluster"
    value = "lattice.system.${var.system_id}"

    propagate_at_launch = true
  }

  tag {
    key   = "Name"
    value = "lattice.system.${var.system_id}.${var.name}"

    propagate_at_launch = true
  }
}
