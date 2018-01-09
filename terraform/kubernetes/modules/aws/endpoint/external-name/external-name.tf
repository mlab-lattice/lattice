###############################################################################
# Variables

variable "region" {}

variable "zone_id" {}
variable "name" {}
variable "external_name" {}

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
  type    = "CNAME"
  ttl     = "60"
  records = ["${var.external_name}"]
}
