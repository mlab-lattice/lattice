###############################################################################
# Variables

variable "region" {}

variable "zone_id" {}
variable "name" {}
variable "ip" {}

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
  type    = "A"
  ttl     = "60"
  records = ["${var.ip}"]
}
