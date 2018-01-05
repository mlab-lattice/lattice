###############################################################################
# Variables

variable "region" {}

variable "cluster_id" {}
variable "system_id" {}
variable "vpc_id" {}
variable "autoscaling_group_name" {}
variable "subnet_ids" {}

variable "service_id" {}
variable "service_security_group_id" {}

# The port_numbers variable maps the port that the load balancer should expose to
# the port on the autoscaling group that it should target.
variable "ports" {
  type = "map"
}

###############################################################################
# Provider

provider "aws" {
  region = "${var.region}"
}

###############################################################################
# ALB

resource "aws_alb" "service" {
  security_groups = ["${aws_security_group.service_lb.id}"]
  subnets         = ["${split(",", var.subnet_ids)}"]
}

# For each port in ports, create a new aws_alb_target_group that targets the
# port exposed on the autoscaling group.
resource "aws_alb_target_group" "service" {
  count = "${length(var.ports)}"

  vpc_id = "${var.vpc_id}"
  port   = "${element(values(var.ports), count.index)}"

  # FIXME: switch to HTTPS when supported
  protocol = "HTTP"

  # FIXME: add health checks
}

# For each port in ports, create a new aws_alb_listener that exposes the
# relevant port and forwards it to the relevant target group.
resource "aws_alb_listener" "service" {
  count = "${length(var.ports)}"

  load_balancer_arn = "${aws_alb.service.arn}"
  port              = "${element(keys(var.ports), count.index)}"

  "default_action" {
    target_group_arn = "${element(aws_alb_target_group.service.*.arn, count.index)}"
    type             = "forward"
  }
}

resource "aws_autoscaling_attachment" "service" {
  count = "${length(var.ports)}"

  autoscaling_group_name = "${var.autoscaling_group_name}"
  alb_target_group_arn   = "${element(aws_alb_target_group.service.*.arn, count.index)}"
}

###############################################################################
# Security group

resource "aws_security_group" "service_lb" {
  vpc_id = "${var.vpc_id}"

  tags {
    KubernetesCluster = "lattice.${var.cluster_id}"
    Name              = "lattice.${var.cluster_id}.system.${var.system_id}.service.${var.service_id}"
  }
}

resource "aws_security_group_rule" "service_allow_ingress_from_lb" {
  count = "${length(var.ports)}"

  security_group_id = "${var.service_security_group_id}"

  type                     = "ingress"
  from_port                = "${element(values(var.ports), count.index)}"
  to_port                  = "${element(values(var.ports), count.index)}"
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.service_lb.id}"
}

resource "aws_security_group_rule" "service_allow_egress_to_lb" {
  count = "${length(var.ports)}"

  security_group_id = "${var.service_security_group_id}"

  type                     = "egress"
  from_port                = "${element(values(var.ports), count.index)}"
  to_port                  = "${element(values(var.ports), count.index)}"
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.service_lb.id}"
}

resource "aws_security_group_rule" "lb_allow_ingress" {
  count = "${length(var.ports)}"

  security_group_id = "${aws_security_group.service_lb.id}"

  type        = "ingress"
  from_port   = "${element(keys(var.ports), count.index)}"
  to_port     = "${element(keys(var.ports), count.index)}"
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "lb_allow_egress_to_service" {
  count = "${length(var.ports)}"

  security_group_id = "${aws_security_group.service_lb.id}"

  type                     = "egress"
  from_port                = "${element(values(var.ports), count.index)}"
  to_port                  = "${element(values(var.ports), count.index)}"
  protocol                 = "tcp"
  source_security_group_id = "${var.service_security_group_id}"
}
