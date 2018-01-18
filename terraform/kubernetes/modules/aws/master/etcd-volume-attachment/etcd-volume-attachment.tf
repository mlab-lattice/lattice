###############################################################################
# Variables

variable "cluster_id" {}
variable "region" {}
variable "name" {}
variable "instance_id" {}
variable "device_name" {}

###############################################################################
# Data

data "aws_ebs_volume" "etcd_data_volume" {
  most_recent = true

  filter {
    name   = "tag:Name"
    values = ["lattice.${var.cluster_id}.master-${var.name}-etcd"]
  }
}

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
# EBS

resource "aws_volume_attachment" "etcd_data_volume_attachment" {
  device_name = "${var.device_name}"
  volume_id   = "${data.aws_ebs_volume.etcd_data_volume.id}"
  instance_id = "${var.instance_id}"
}
