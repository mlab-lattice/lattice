###############################################################################
# Variables

variable "region" {}

variable "zone_id" {}
variable "type" {}
variable "name" {}
variable "value" {}

###############################################################################
# Provider

provider "aws" {
  region = "${var.region}"
}

###############################################################################
# Record

resource "aws_route53_record" "endpoint" {
  zone_id = "${var.zone_id}"
  name    = "${var.name}"
  type    = "${var.type}"
  ttl     = "60"
  records = ["${var.value}"]
}
