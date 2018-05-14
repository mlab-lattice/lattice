###############################################################################
# Variables

variable "region" {}

variable "lattice_id" {}
variable "system_id" {}
variable "vpc_id" {}
variable "subnet_ids" {}

variable "name" {}

# The autoscaling_group_security_group_ids variable maps the name of an autoscaling
# group that the load balancer should attach to to the id of its security group.
variable "autoscaling_group_security_group_ids" {
  type = "map"
}

# The ports variable maps the port that the load balancer should expose to
# the port on the autoscaling group that it should target.
variable "ports" {
  type = "map"
}

###############################################################################
# Output
#

output "dns_name" {
  value = "${aws_alb.load_balancer.dns_name}"
}

###############################################################################
# Provider

provider "aws" {
  region = "${var.region}"
}

###############################################################################
# ALB

resource "aws_alb" "load_balancer" {
  security_groups = ["${aws_security_group.load_balancer_lb.id}"]
  subnets         = ["${split(",", var.subnet_ids)}"]

  tags {
    KubernetesCluster = "lattice.${var.lattice_id}"
    Name              = "lattice.${var.lattice_id}.system.${var.system_id}.load-balancer.${var.name}"
  }
}

# For each port in ports, create a new aws_alb_target_group that targets the
# port exposed on the autoscaling group.
resource "aws_alb_target_group" "load_balancer" {
  count = "${length(var.ports)}"

  vpc_id = "${var.vpc_id}"
  port   = "${element(values(var.ports), count.index)}"

  # FIXME: switch to HTTPS when supported
  protocol = "HTTP"

  # FIXME: add health checks

  tags {
    KubernetesCluster = "lattice.${var.lattice_id}"
    Name              = "lattice.${var.lattice_id}.system.${var.system_id}.load-balancer.${var.name}"
  }
}

# For each port in ports, create a new aws_alb_listener that exposes the
# relevant port and forwards it to the relevant target group.
resource "aws_alb_listener" "load_balancer" {
  count = "${length(var.ports)}"

  load_balancer_arn = "${aws_alb.load_balancer.arn}"
  port              = "${element(keys(var.ports), count.index)}"

  "default_action" {
    target_group_arn = "${element(aws_alb_target_group.load_balancer.*.arn, count.index)}"
    type             = "forward"
  }
}

# For each autoscaling group, for each port in ports, attach the autoscaling group
# to the target group for the port.
resource "aws_autoscaling_attachment" "load_balancer" {
  count = "${length(var.ports) * length(var.autoscaling_group_security_group_ids)}"

  autoscaling_group_name = "${element(keys(var.autoscaling_group_security_group_ids), count.index % length(var.autoscaling_group_security_group_ids))}"
  alb_target_group_arn   = "${element(aws_alb_target_group.load_balancer.*.arn, count.index % length(var.ports))}"
}

###############################################################################
# Security group

resource "aws_security_group" "load_balancer_lb" {
  vpc_id = "${var.vpc_id}"

  tags {
    KubernetesCluster = "lattice.${var.lattice_id}"
    Name              = "lattice.${var.lattice_id}.system.${var.system_id}.load-balancer.${var.name}"
  }
}

# For each autoscaling group, for each port in ports, add the load balancer to the
# autoscaling group's ingress allow rules for the port
resource "aws_security_group_rule" "load_balancer_allow_ingress_asg_from_lb" {
  count = "${length(var.ports) * length(var.autoscaling_group_security_group_ids)}"

  security_group_id = "${element(values(var.autoscaling_group_security_group_ids), count.index % length(var.autoscaling_group_security_group_ids))}"

  type                     = "ingress"
  from_port                = "${element(values(var.ports), count.index % length(var.ports))}"
  to_port                  = "${element(values(var.ports), count.index % length(var.ports))}"
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.load_balancer_lb.id}"
}

# For each autoscaling group, for each port in ports, add the load balancer to the
# autoscaling group's egress allow rules for the port.
resource "aws_security_group_rule" "load_balancer_allow_egress_asg_to_lb" {
  count = "${length(var.ports) * length(var.autoscaling_group_security_group_ids)}"

  security_group_id = "${element(values(var.autoscaling_group_security_group_ids), count.index % length(var.autoscaling_group_security_group_ids))}"

  type                     = "egress"
  from_port                = "${element(values(var.ports), count.index % length(var.ports))}"
  to_port                  = "${element(values(var.ports), count.index % length(var.ports))}"
  protocol                 = "tcp"
  source_security_group_id = "${aws_security_group.load_balancer_lb.id}"
}

# Add each port to the load balancer's ingress allow rules.
resource "aws_security_group_rule" "lb_allow_ingress" {
  count = "${length(var.ports)}"

  security_group_id = "${aws_security_group.load_balancer_lb.id}"

  type        = "ingress"
  from_port   = "${element(keys(var.ports), count.index)}"
  to_port     = "${element(keys(var.ports), count.index)}"
  protocol    = "tcp"
  cidr_blocks = ["0.0.0.0/0"]
}

# For each autoscaling group, for each port in ports, add the autoscaling group
# to the load balancer's egress allow rules for the port.
resource "aws_security_group_rule" "lb_allow_egress_to_load_balancer" {
  count = "${length(var.ports) * length(var.autoscaling_group_security_group_ids)}"

  security_group_id = "${aws_security_group.load_balancer_lb.id}"

  type                     = "egress"
  from_port                = "${element(values(var.ports), count.index)}"
  to_port                  = "${element(values(var.ports), count.index)}"
  protocol                 = "tcp"
  source_security_group_id = "${element(values(var.autoscaling_group_security_group_ids), count.index % length(var.autoscaling_group_security_group_ids))}"
}
