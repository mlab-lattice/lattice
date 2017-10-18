###############################################################################
# Variables

variable "system_id" {}
variable "region" {}
variable "name" {}
variable "route53_private_zone_id" {}
variable "instance_private_ip" {}

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
# Route53

resource "aws_route53_record" "kube_masters_entry" {
  zone_id = "${var.route53_private_zone_id}"
  name    = "kube-masters"

  type    = "A"
  ttl     = 60
  records = ["${var.instance_private_ip}"]
}

resource "aws_route53_record" "kube_master_entry" {
  zone_id = "${var.route53_private_zone_id}"
  name    = "kube-master-${var.name}"

  type    = "A"
  ttl     = 60
  records = ["${var.instance_private_ip}"]
}
